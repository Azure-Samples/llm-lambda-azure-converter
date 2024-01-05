package lats

import (
	"fmt"
	"io"
	"os"

	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/models"
)

const (
	GoSimpleChatInstruction         = "You are an AI that only responds with Go code, NOT ENGLISH. You will be given a function signature and its docstring by the user. Write your full implementation (restate the function signature)."
	GoReflectionChatInstruction     = "You are an AI Go assistant. You will be given your past function implementation, a series of unit tests, and a hint to change the implementation appropriately. Write your full implementation (restate the function signature)."
	GoSelfReflectionChatInstruction = "You are a Go programming assistant. You will be given a function implementation and a series of unit tests. Your goal is to write a few sentences to explain why your implementation is wrong as indicated by the tests. You will need this as a hint when you try again later. Only provide the few sentence description in your answer, not the implementation."
	GoTestGenerationChatInstruction = "You are a Go programming assistant, an AI coding assistant that can write unique, diverse, and intuitive unit tests for functions given the signature and docstring."
)

var (
	GoReflectionFewShotAdd  string
	GoTestGenerationFewShot string
	GoSelfReflectionFewShot string
)

func init() {
	GoReflectionFewShotAdd = readTxt("./prompt/GoReflectionFewShotAdd.md")
	GoTestGenerationFewShot = readTxt("./prompt/GoTestGenerationFewShot.md")
	GoSelfReflectionFewShot = readTxt("./prompt/GoSelfReflectionFewShot.md")
}

func readTxt(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Could not load file %s\n", filename)
		return ""
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Could not load file %s\n", filename)
		return ""
	}
	return string(bytes)
}

type goGenerator struct {
	llm models.LLM
}

func NewGoGenerator(llm models.LLM) models.Generator {
	return &goGenerator{
		llm: llm,
	}
}

func (g *goGenerator) GenerateCode(code string) (string, error){
	return "", nil
}

func (g *goGenerator) GenerateTests(code string) ([]string, error){
	return nil, nil
}

func (g *goGenerator) GenerateReflection(code string) (string, error){
	return "", nil
}
