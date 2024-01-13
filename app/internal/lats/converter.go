package lats

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/models"
	"github.com/rs/zerolog"
)

type converter struct {
	generator     models.Generator
	executor      models.Executor
	maxIterations int
	maxChildren   int
	logger        zerolog.Logger
}

func NewConverter(generator models.Generator, executor models.Executor, config LatsConfig) models.Converter {
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

func generateStatistics(node models.Node, startTime time.Time) *models.ConverterStatistics {
	return &models.ConverterStatistics{
		TotalIterations: node.Iteration,
		SelectedNode:    node.Id,
		TotalTime:       time.Since(startTime),
		Found:           node.Score == 1,
	}
}

func (m *converter) Convert(ctx context.Context, code string, originalTests []string, generateTests bool) (*string, *models.ConverterStatistics, error) {
	startTime := time.Now()
	rootNode, err := m.generateNode(ctx, code, nil, originalTests, generateTests)
	if err != nil {
		return nil, nil, err
	}

	if rootNode.Score == 1 {
		return &rootNode.Code, generateStatistics(*rootNode, startTime), nil
	}

	currentIteration := 0
	bestNode := rootNode
	currentNode := rootNode
	for currentIteration <= m.maxIterations {
		m.logger.Debug().Msgf("Iteration %d", currentIteration)

		for len(currentNode.ChildNodes) < m.maxChildren {
			m.logger.Debug().Msgf("Generating child node %d", len(currentNode.ChildNodes)+1)

			childNode, err := m.generateNode(ctx, code, currentNode, originalTests, generateTests)
			if err != nil {
				return nil, nil, err
			}

			if childNode.Score == 1 {
				return &childNode.Code, generateStatistics(*childNode, startTime), nil
			}

			if childNode.Score >= bestNode.Score {
				bestNode = childNode
			}
			m.logger.Debug().Msgf("Best score:\n%f", bestNode.Score)
		}
		currentNode = bestNode
		currentIteration++
	}

	m.logger.Debug().Msgf("Could not find a better solution after %d iterations", m.maxIterations)
	m.logger.Debug().Msgf("Returning code with best score %f;\n%+v", bestNode.Score, bestNode.Code)
	return &bestNode.Code, generateStatistics(*bestNode, startTime), nil
}

func (m *converter) generateNode(ctx context.Context, code string, parentNode *models.Node, originalTests []string, generateTests bool) (*models.Node, error) {
	nodeIteration := 0
	nodeId := "0"
	if parentNode != nil {
		nodeIteration = parentNode.Iteration + 1
		nodeId = fmt.Sprintf("%s.%d", parentNode.Id, len(parentNode.ChildNodes))
	}

	var generatedCode *string
	var err error
	if parentNode == nil {
		generatedCode, err = m.generator.GenerateCode(ctx, code)
	} else {
		generatedCode, err = m.generator.GenerateCodeWithReflection(ctx, code, parentNode.Code, parentNode.Feedback, parentNode.SelfReflection)
	}
	if err != nil {
		return nil, fmt.Errorf("there was an error generating code on iteration %d, node %s: %v", nodeIteration, nodeId, err)
	}
	m.logger.Debug().Msgf("Generated code:\n%s", *generatedCode)

	tests := make([]string, 0)
	tests = append(tests, originalTests...)
	if generateTests {
		generatedTests, err := m.generator.GenerateTests(ctx, code, *generatedCode)
		if err != nil {
			return nil, fmt.Errorf("there was an error generating tests on iteration %d, node %s: %v", nodeIteration, nodeId, err)
		}
		m.logger.Debug().Msgf("Generated tests:\n%s", strings.Join(generatedTests, "\n\n"))
		tests = append(tests, generatedTests...)
	}

	result, err := m.executor.Execute(*generatedCode, tests)
	if err != nil {
		return nil, fmt.Errorf("there was an error running/testing code on iteration %d, node %s: %v", nodeIteration, nodeId, err)
	}
	m.logger.Debug().Msgf("Execution result:\n%+v", result)

	var selfReflection string
	if !result.IsPassing {
		reflection, err := m.generator.GenerateSelfReflection(ctx, *generatedCode, result.Feedback)
		if err != nil {
			return nil, fmt.Errorf("there was an error generating reflection on iteration %d, node %s: %v", nodeIteration, nodeId, err)
		}
		selfReflection = *reflection
		m.logger.Debug().Msgf("Generated self-reflection:\n%s", *reflection)
	}

	newNode := &models.Node{
		Iteration:      nodeIteration,
		Id:             nodeId,
		Code:           *generatedCode,
		Feedback:       result.Feedback,
		SelfReflection: selfReflection,
		Score:          result.Score,
		ParentNode:     parentNode,
		ChildNodes:     make([]*models.Node, 0),
	}
	if parentNode != nil {
		parentNode.ChildNodes = append(parentNode.ChildNodes, newNode)
	}

	return newNode, nil
}
