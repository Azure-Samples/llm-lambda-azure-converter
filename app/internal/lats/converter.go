package lats

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/models"
	"github.com/rs/zerolog"
)

type Converter interface {
	Convert(ctx context.Context, code string, tests []string, generateTests bool) (*string, bool, error)
}

type converter struct {
	generator     models.Generator
	executor      models.Executor
	maxIterations int
	maxChildren   int
	logger        zerolog.Logger
}

func NewConverter(generator models.Generator, executor models.Executor, config LatsConfig) Converter {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		With().Timestamp().Caller().Logger()
	return &converter{
		generator:     generator,
		executor:      executor,
		maxIterations: config.ConverterMaxIterations,
		maxChildren:   config.ConverterMaxChildren,
		logger:        logger,
	}
}

func (m *converter) Convert(ctx context.Context, code string, originalTests []string, generateTests bool) (*string, bool, error) {

	generatedCode, err := m.generator.GenerateCode(ctx, code)
	if err != nil {
		return nil, false, err
	}
	m.logger.Debug().Msgf("Generated code:\n%s", *generatedCode)

	var generatedTests []string
	if generateTests {
		signature, err := m.generator.QueryFuncSignature(ctx, *generatedCode)
		if err != nil {
			return nil, false, err
		}
		m.logger.Debug().Msgf("Generated signature:\n%s", *signature)

		generatedTests, err = m.generator.GenerateTests(ctx, *signature, code)
		if err != nil {
			return nil, false, err
		}
		m.logger.Debug().Msgf("Generated tests:\n%s", strings.Join(generatedTests, "\n\n"))
	} else {
		generatedTests = originalTests
	}

	result, err := m.executor.Execute(*generatedCode, generatedTests)
	if err != nil {
		return nil, false, err
	}
	m.logger.Debug().Msgf("Execution result:\n%+v", result)

	if result.IsPassing {
		finalResult, err := m.executor.Execute(*generatedCode, originalTests)
		if err == nil && finalResult.IsPassing {
			return generatedCode, true, nil
		}
		result = finalResult
		m.logger.Debug().Msgf("Result is passing but fails on original tests. Using original tests for feedback:\n%+v", finalResult)
	}

	selfReflection, err := m.generator.GenerateSelfReflection(ctx, *generatedCode, result.Feedback)
	if err != nil {
		return nil, false, err
	}
	m.logger.Debug().Msgf("Generated self-reflection:\n%s", *selfReflection)

	currentNode := &models.Node{
		Code:           *generatedCode,
		Feedback:       result.Feedback,
		SelfReflection: *selfReflection,
		Score:          result.Score,
		ParentNode:     nil,
		ChildNodes:     make([]*models.Node, 0),
	}
	currentIteration := 1
	bestNode := currentNode
	for currentIteration <= m.maxIterations {
		m.logger.Debug().Msgf("Iteration %d", currentIteration)

		for len(currentNode.ChildNodes) < m.maxChildren {
			m.logger.Debug().Msgf("Generating child node %d", len(currentNode.ChildNodes)+1)

			newNode, err := m.generateNode(ctx, code, currentNode)
			if err != nil {
				return nil, false, err
			}

			if newNode.Score == 1 {
				finalResult, err := m.executor.Execute(newNode.Code, originalTests)
				if err != nil && finalResult.IsPassing {
					return generatedCode, true, nil
				}
				newNode.Score = finalResult.Score
				newNode.Feedback = finalResult.Feedback
				m.logger.Debug().Msgf("Result is passing but fails on original tests. Using original tests for feedback:\n%+v", finalResult)
			}

			reflection, err := m.generator.GenerateSelfReflection(ctx, newNode.Code, newNode.Feedback)
			if err != nil {
				return nil, false, err
			}
			newNode.SelfReflection = *reflection

			if newNode.Score >= bestNode.Score {
				bestNode = newNode
			}
		}
		currentNode = bestNode
		currentIteration++
	}

	m.logger.Debug().Msgf("Could not find a better solution after %d iterations", m.maxIterations)
	m.logger.Debug().Msgf("Returning code with best score %f;\n%+v", bestNode.Score, bestNode.Code)
	return &bestNode.Code, false, nil
}

func (m *converter) generateNode(ctx context.Context, code string, parentNode *models.Node) (*models.Node, error) {

	generatedCode, err := m.generator.GenerateCodeWithReflection(ctx, code, parentNode.Code, parentNode.Feedback, parentNode.SelfReflection)
	if err != nil {
		return nil, err
	}

	signature, err := m.generator.QueryFuncSignature(ctx, *generatedCode)
	if err != nil {
		return nil, err
	}

	generatedTests, err := m.generator.GenerateTests(ctx, *signature, code)
	if err != nil {
		return nil, err
	}

	result, err := m.executor.Execute(*generatedCode, generatedTests)
	if err != nil {
		return nil, err
	}

	newNode := &models.Node{
		Code:           *generatedCode,
		Feedback:       result.Feedback,
		SelfReflection: "",
		Score:          result.Score,
		ParentNode:     parentNode,
		ChildNodes:     make([]*models.Node, 0),
	}
	parentNode.ChildNodes = append(parentNode.ChildNodes, newNode)

	return newNode, nil
}
