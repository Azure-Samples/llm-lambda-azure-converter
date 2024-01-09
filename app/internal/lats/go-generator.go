package lats

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/models"
)

const (
	GoSimpleChatInstruction         = "You are an AI that only responds with Go code, NOT ENGLISH. You will be given a function signature and its docstring by the user. Write your full implementation (restate the function signature)."
	GoReflectionChatInstruction     = "You are an AI Go assistant. You will be given your past function implementation, a series of unit tests, and a hint to change the implementation appropriately. Write your full implementation (restate the function signature)."
	GoSelfReflectionChatInstruction = "You are a Go programming assistant. You will be given a function implementation and a series of unit tests. Your goal is to write a few sentences to explain why your implementation is wrong as indicated by the tests. You will need this as a hint when you try again later. Only provide the few sentence description in your answer, not the implementation."
	GoTestGenerationChatInstruction = "You are a Go programming assistant, an AI coding assistant that can write unique, diverse, and intuitive unit tests for functions given the signature and an equivalent code."
	GoSignatureChatInstruction      = "You are an AI Go assistant. You will be given a function implementation, and from it you will extract the handler function signature."
	GoCodeBlockInstruction          = "Use a Go code block to write your response. For example:\n```go\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n```"
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

func (g *goGenerator) GenerateCode(ctx context.Context, code string) (*string, error) {
	systemInstruction := fmt.Sprintf("%s\n%s", GoSimpleChatInstruction, GoCodeBlockInstruction)
	fmt.Println(systemInstruction)

	messages := []models.ChatMessage{
		{Type: models.SystemMessage, Content: systemInstruction},
		{Type: models.UserMessage, Content: code},
	}

	return g.llm.Chat(ctx, messages)
}

func (g *goGenerator) GenerateCodeWithReflection(ctx context.Context, code string, previousResult string, feedback string, selfReflection string) (*string, error) {
	systemInstruction := fmt.Sprintf("%s\n%s", GoReflectionChatInstruction, GoCodeBlockInstruction)
	fmt.Println(systemInstruction)

	messages := []models.ChatMessage{
		{Type: models.SystemMessage, Content: systemInstruction},
		{Type: models.UserMessage, Content: GoReflectionFewShotAdd},
		{Type: models.AssistantMessage, Content: fmt.Sprintf("```go\n%s\n```", previousResult)},
		{Type: models.UserMessage, Content: fmt.Sprintf("[unit test results from previous impl]:\n%s\n\n[reflection on previous impl]:", feedback)},
		{Type: models.AssistantMessage, Content: selfReflection},
		{Type: models.UserMessage, Content: fmt.Sprintf("[improved impl]:\n%s", code)},
	}

	return g.llm.Chat(ctx, messages)
}

func (g *goGenerator) GenerateSelfReflection(ctx context.Context, code string, feedback string) (*string, error) {
	systemInstruction := fmt.Sprintf("%s\n%s", GoSimpleChatInstruction, GoCodeBlockInstruction)
	fmt.Println(systemInstruction)

	messages := []models.ChatMessage{
		{Type: models.SystemMessage, Content: GoSelfReflectionChatInstruction},
		{Type: models.UserMessage, Content: fmt.Sprintf("%s\n\n[function impl]:\n```go\n%s\n```\n\n[unit test results]:\n%s\n\n[self-reflection]:", GoSelfReflectionFewShot, code, feedback)},
	}

	return g.llm.Chat(ctx, messages)
}

func (g *goGenerator) GenerateTests(ctx context.Context, funcSignature string) (*string, error) {
	messages := []models.ChatMessage{
		{Type: models.SystemMessage, Content: fmt.Sprintf("%s\n%s", GoTestGenerationChatInstruction, GoTestGenerationFewShot)},
		{Type: models.UserMessage, Content: fmt.Sprintf("[func signature]:\n%s\n\n[unit tests]:", funcSignature)},
	}

	return g.llm.Chat(ctx, messages)
}

func (g *goGenerator) QueryFuncSignature(ctx context.Context, code string) (*string, error) {
	messages := []models.ChatMessage{
		{Type: models.SystemMessage, Content: fmt.Sprintf("%s\n%s", GoSignatureChatInstruction, GoCodeBlockInstruction)},
		{Type: models.UserMessage, Content: code},
	}

	return g.llm.Chat(ctx, messages)
}
