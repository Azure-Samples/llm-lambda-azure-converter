/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/lats"
	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	evalDataPath string
	evalLanguage string
)

type EvaluationCase struct {
	Id          string   `json:"id"`
	ProjectPath string   `json:"projectPath"`
	MainFile    string   `json:"mainFile"`
	TargetPath  string   `json:"targetPath"`
	TestPaths   []string `json:"testPaths"`
}

type EvaluationResult struct {
	Id              string
	Found           bool
	TotalIterations int
	TotalTime       time.Duration
	TotalAttempts   int
}

type EvaluationError struct {
	Id    string
	Error string
}

var evaluateCmd = &cobra.Command{
	Use:   "evaluate",
	Short: "Evaluate converter performance",
	Long:  `The converter will run for multiple inputs and evaluate its performance based on the results.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Evaluate called")

		config := lats.NewLatsConfig(*viper.GetViper())
		llm, err := lats.NewOpenAIChat(*config)
		if err != nil {
			return fmt.Errorf("error creating the LLM: %v", err)
		}

		var converter models.Converter
		switch evalLanguage {
		case "go":
			goExecutor := lats.NewGoExecutor()
			goGenerator := lats.NewGoGenerator(llm, promptsDir)
			converter = lats.NewConverter(goGenerator, goExecutor, *config)
		default:
			return fmt.Errorf("unsupported language: %s", evalLanguage)
		}

		cases, err := readEvalData(evalDataPath)
		if err != nil {
			return fmt.Errorf("could not read the evaluation data file: %s", err.Error())
		}
		if len(cases) == 0 {
			return fmt.Errorf("there are no cases in the data file: %s", err.Error())
		}

		results := make([]EvaluationResult, 0)
		errors := make([]EvaluationError, 0)
		for _, evalCase := range cases {
			code, err := readFile(path.Join(evalCase.ProjectPath, evalCase.MainFile))
			if err != nil {
				errMsg := fmt.Sprintf("could not read the code file: %s", err.Error())
				fmt.Println(errMsg)
				errors = append(errors, EvaluationError{
					Id:    evalCase.Id,
					Error: errMsg,
				})
				continue
			}
			code = "```go\n" + code + "\n```\n"

			tests := make([]string, 0)
			for _, path := range evalCase.TestPaths {
				test, err := readFile(path)
				if err != nil {
					return fmt.Errorf("could not read the test file: %s", err.Error())
				}
				test = "```go\n" + test + "\n```\n"
				tests = append(tests, test)
			}

			response, err := converter.Convert(
				context.Background(),
				code,
				tests,
				models.WithGenerateTests(generateTests),
				models.WithProjectPath(evalCase.ProjectPath),
				models.WithTargetPath(evalCase.TargetPath),
				models.WithMainFile(evalCase.MainFile),
			)
			if err != nil {
				fmt.Printf("there was an error converting the code for the case '%s': %s", evalCase.Id, err.Error())
				continue
			}

			fmt.Printf("Case '%s':\n", evalCase.Id)
			if response.Found {
				fmt.Printf("Found a solution in %d iterations and %d attempts\n", response.TotalIterations, response.TotalAttempts)
			} else {
				fmt.Printf("Couldn't find a solution after %d iterations and %d attempts\n", response.TotalIterations, response.TotalAttempts)
			}
	
			fmt.Printf("Total time: %s\n", response.TotalTime.String())
			fmt.Printf("Total attempts: %d\n", response.TotalAttempts)

			results = append(results, EvaluationResult{
				Id:              evalCase.Id,
				Found:           response.Found,
				TotalIterations: response.TotalIterations,
				TotalAttempts:   response.TotalAttempts,
				TotalTime:       response.TotalTime,
			})
		}

		fmt.Println("********************************************")
		fmt.Println("Aggregate Results:")
		fmt.Println("********************************************")
		fmt.Println("id                   | found  | iterations | attempts | time")
		fmt.Println("---------------------|--------|------------|----------|------------")
		for _, result := range results {
			fmt.Printf("%-20s | %-6t | %-10d | %-8d | %v \n",
				result.Id,
				result.Found,
				result.TotalIterations,
				result.TotalAttempts,
				result.TotalTime.Truncate(time.Second),
			)
		}
		fmt.Println("********************************************")
		fmt.Println("Aggregate Errors:")
		fmt.Println("********************************************")
		fmt.Println("id                 | error")
		fmt.Println("-------------------|------")
		for _, err := range errors {
			fmt.Printf("%-20s | %s\n", err.Id, err.Error)
		}
		fmt.Println("********************************************")
		fmt.Println("Summary:")
		fmt.Println("********************************************")
		fmt.Printf("Total Cases: %d\n", len(cases))
		fmt.Printf("Total Time: %v\n", totalTime(results))
		fmt.Printf("Found solutions: %v\n", averageFound(results))
		fmt.Printf("Average attempts: %2.f\n", averageAttempts(results))
		fmt.Printf("Average iterations: %2.f\n", averageIterations(results))
		fmt.Printf("Average time: %v\n", averageTime(results))
		fmt.Println("********************************************")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(evaluateCmd)

	evaluateCmd.Flags().StringVarP(&evalDataPath, "data-path", "e", "", "The path to the jsonl file with the evaluation cases")
	evaluateCmd.Flags().BoolVarP(&generateTests, "generate-tests", "g", true, "Generate tests for the conversion")
	evaluateCmd.Flags().StringVarP(&evalLanguage, "language", "l", "go", "The language of the lambda functions")
	evaluateCmd.MarkFlagRequired("data-path")
}

func readEvalData(path string) ([]EvaluationCase, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("failed to close file: %v", err)
		}
	}()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	evalCases := make([]EvaluationCase, 0)
	for _, line := range lines {
		evalCase := EvaluationCase{}
		json.Unmarshal([]byte(line), &evalCase)
		evalCases = append(evalCases, evalCase)
	}

	return evalCases, nil
}

func averageFound(results []EvaluationResult) float64 {
	if len(results) == 0 {
		return 0
	}

	sum := 0
	for _, result := range results {
		if result.Found {
			sum++
		}
	}
	return float64(sum) / float64(len(results))
}

func averageIterations(results []EvaluationResult) float64 {
	if len(results) == 0 {
		return 0
	}

	sum := 0
	for _, result := range results {
		sum += result.TotalIterations
	}
	return float64(sum) / float64(len(results))
}

func averageAttempts(results []EvaluationResult) float64 {
	if len(results) == 0 {
		return 0
	}

	sum := 0
	for _, result := range results {
		sum += result.TotalAttempts
	}
	return float64(sum) / float64(len(results))
}

func totalTime(results []EvaluationResult) time.Duration {
	if len(results) == 0 {
		return 0
	}

	sum := time.Duration(0)
	for _, result := range results {
		sum += result.TotalTime
	}
	return time.Duration(sum)
}

func averageTime(results []EvaluationResult) time.Duration {
	if len(results) == 0 {
		return 0
	}

	sum := time.Duration(0)
	for _, result := range results {
		sum += result.TotalTime
	}
	return sum / time.Duration(len(results))
}
