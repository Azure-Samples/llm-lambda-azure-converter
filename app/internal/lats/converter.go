package lats

import (
	"context"
	"fmt"
	"os"
	"path"
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

func buildResponse(node models.Node, startTime time.Time, attempts int) *models.ConverterResponse {
	return &models.ConverterResponse{
		Code:            node.Code,
		Tests:           node.Tests,
		TotalIterations: node.Iteration,
		TotalAttempts:   attempts,
		SelectedNode:    node.Id,
		TotalTime:       time.Since(startTime),
		Found:           node.Score == 1,
	}
}

func (m *converter) Convert(ctx context.Context, code string, originalTests []string, options ...models.ConverterOption) (*models.ConverterResponse, error) {
	startTime := time.Now()

	converterOptions := &models.ConverterOptions{}
	for _, option := range options {
		option(converterOptions)
	}

	rootNode, err := m.generateNode(ctx, code, nil, originalTests, *converterOptions)
	if err != nil {
		return nil, err
	}
	attempts := 1

	if rootNode.Score == 1 {
		return buildResponse(*rootNode, startTime, attempts), nil
	}

	currentIteration := 0
	bestNode := rootNode
	currentNode := rootNode
	for currentIteration <= m.maxIterations {
		m.logger.Debug().Msgf("Iteration %d", currentIteration)

		for len(currentNode.ChildNodes) < m.maxChildren {
			m.logger.Debug().Msgf("Generating child node %d", len(currentNode.ChildNodes)+1)

			childNode, err := m.generateNode(ctx, code, currentNode, originalTests, *converterOptions)
			if err != nil {
				return nil, err
			}
			attempts++

			if childNode.Score == 1 {
				return buildResponse(*childNode, startTime, attempts), nil
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
	return buildResponse(*bestNode, startTime, attempts), nil
}

func (m *converter) generateNode(ctx context.Context, code string, parentNode *models.Node, originalTests []string, options models.ConverterOptions) (*models.Node, error) {
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
	if options.GenerateTests {
		if parentNode == nil {
			tests = append(tests, originalTests...)
			generatedTests, err := m.generator.GenerateTests(ctx, code, *generatedCode)
			if err != nil {
				return nil, fmt.Errorf("there was an error generating tests on iteration %d, node %s: %v", nodeIteration, nodeId, err)
			}
			tests = append(tests, generatedTests...)
			m.logger.Debug().Msgf("Generated tests:\n%s", strings.Join(generatedTests, "\n\n"))
		} else {
			tests = parentNode.Tests
		}
	}

	result, err := m.executor.Execute(
		*generatedCode,
		tests,
		models.WithExProjectPath(options.ProjectPath),
		models.WithExExTargetPath(options.TargetPath),
		models.WithExFilename(strings.Split(path.Base(options.MainFile), ".")[0]),
		models.WithExCreateProject(options.CreateProject))

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

		implementationIsGood, err := m.generator.QueryImplementationIsGood(ctx, *reflection)
		if err != nil {
			return nil, fmt.Errorf("there was an error querying the reflection on iteration %d, node %s: %v", nodeIteration, nodeId, err)
		}
		m.logger.Debug().Msgf("The implementation is good, problem is in the tests:\n%s", *implementationIsGood)

		if strings.Contains(strings.ToLower(*implementationIsGood), "yes") {
			m.logger.Debug().Msgf("Running execution again without generated tests")
			tests = originalTests

			result, err = m.executor.Execute(
				*generatedCode,
				originalTests,
				models.WithExProjectPath(options.ProjectPath),
				models.WithExExTargetPath(options.TargetPath),
				models.WithExFilename(strings.Split(path.Base(options.MainFile), ".")[0]),
				models.WithExCreateProject(options.CreateProject))

			if err != nil {
				return nil, fmt.Errorf("there was an error running/testing code on iteration %d, node %s: %v", nodeIteration, nodeId, err)
			}
			m.logger.Debug().Msgf("Second execution result:\n%+v", result)

			if !result.IsPassing {
				reflection, err := m.generator.GenerateSelfReflection(ctx, *generatedCode, result.Feedback)
				if err != nil {
					return nil, fmt.Errorf("there was an error generating reflection on iteration %d, node %s: %v", nodeIteration, nodeId, err)
				}
				selfReflection = *reflection
				m.logger.Debug().Msgf("Second generated self-reflection:\n%s", *reflection)
			}
		}
	}

	newNode := &models.Node{
		Iteration:      nodeIteration,
		Id:             nodeId,
		Code:           *generatedCode,
		Tests:          tests,
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
