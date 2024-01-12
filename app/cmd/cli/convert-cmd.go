/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/lats"
	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	promptsDir = "../../internal/lats/prompts"
)

var (
	language      string
	generateTests bool
	codePath      string
	testsPath     []string
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert a lambda function to an azure function",
	Long: `The converter will take a lambda function and go through a series of 
	iterations where it will try to convert the lambda code, generate tests to check it's right,
	and in case the code doesn't compile or it doesn't pass the tests, it will do a self reflection
	and based on it and the tests feedback it will try to fix the code and generate the tests again.
	This will happen until the code compiles and passes the tests, or the maximum number of iterations
	is reached.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("convert called")

		config := lats.NewLatsConfig(*viper.GetViper())
		llm, err := lats.NewOpenAIChat(*config)
		if err != nil {
			return fmt.Errorf("error creating the LLM: %v", err)
		}

		var converter models.Converter
		switch language {
		case "go":
			goExecutor := lats.NewGoExecutor()
			goGenerator := lats.NewGoGenerator(llm, promptsDir)
			converter = lats.NewConverter(goGenerator, goExecutor, *config)
		default:
			return fmt.Errorf("unsupported language: %s", language)
		}

		code, err := readFile(codePath)
		if err != nil {
			return fmt.Errorf("could not read the code file: %s", err.Error())
		}
		code = "```go\n" + code + "\n```\n"

		tests := make([]string, 0)
		for _, path := range testsPath {
			test, err := readFile(path)
			if err != nil {
				return fmt.Errorf("could not read the test file: %s", err.Error())
			}
			test = "```go\n" + test + "\n```\n"
			tests = append(tests, test)
		}

		generatedCode, info, err := converter.Convert(context.Background(), code, tests, generateTests)

		if err != nil {
			return fmt.Errorf("there was an error converting the code: %s", err.Error())

		}

		if info.Found {
			fmt.Printf("Found a solution in %d iterations\n", info.TotalIterations)
		} else {
			fmt.Printf("Couldn't find a solution after %d iterations\n", info.TotalIterations)
		}
		fmt.Printf("Total time: %s\n", info.TotalTime.String())
		fmt.Printf("Showing code for node %s\n", info.SelectedNode)
		fmt.Println("")
		fmt.Println(*generatedCode)

		return nil
	},
}

func readFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("error opening the file: %s", err.Error())
	}
	code, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("error reading the file: %s", err.Error())
	}
	return string(code), nil
}

func init() {
	rootCmd.AddCommand(convertCmd)

	convertCmd.Flags().StringVarP(&language, "language", "l", "go", "The language of the lambda function")
	convertCmd.Flags().BoolVarP(&generateTests, "generate-tests", "g", true, "Generate tests for the conversion")
	convertCmd.Flags().StringVarP(&codePath, "code-path", "c", "", "The path to the lambda function code")
	convertCmd.Flags().StringArrayVarP(&testsPath, "test-path", "t", []string{}, "The path to the lambda function tests")
	convertCmd.MarkFlagRequired("code-path")
}
