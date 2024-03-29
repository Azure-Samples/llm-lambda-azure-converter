package lats

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/models"
)

const (
	goSimpleChatInstruction             = "You are an AI Go assistant that converts code implementations. you only respond with Go code, NOT ENGLISH. You will be given a lambda function code. Rewrite the code without using lambda code and using a GinGonic server instead.\nIf you can change the handler function, change it, but keep the same name.\nDo not by any chance use lambda code."
	goReflectionChatInstruction         = "You are an AI Go assistant that converts code implementations. You will be given a lambda function code, your past function conversion, a series of unit tests, and a hint to change the implementation appropriately. Write your full implementation, Rewrite the code without using lambda code and using a GinGonic server instead.\nIf you can change the handler function, change it, but keep the same name.\nDo not by any chance use lambda code."
	goSelfReflectionChatInstruction     = "You are a Go programming assistant. You will be given a function implementation and a series of unit tests. Your goal is to write a few sentences to explain why your implementation is wrong as indicated by the tests. You will need this as a hint when you try again later. Only provide the few sentence description in your answer, not the implementation."
	goTestGenerationChatInstruction     = "You are a Go programming assistant, an AI coding assistant that can write unique, diverse, and intuitive unit tests for functions. You will be given a Go AWS Lambda function, that is being converted to a GinGonic http server. Your job is to generate a comprehensive set of tests to ensure its functionality remains consistent. The tests should cover all major functionality of the function, including error handling, input validation, and expected output. Generate self-contained tests that do not rely on external resources or dependencies. Do not include the main function in your tests."
	goTestGenerationHumanInstruction    = "Here is the Go code for the AWS Lambda function: \n%s\n\nHere is the Go code for the GinGonic http server:\n%s\n"
	goYesNoChatInstruction              = "You are an AI Go assistant. You will be given a reflection about an implementation, and from it you will answer a question just with 'yes' or 'no'. Do not answer with any different words, just either 'yes' or 'no'."
	goImplementationGoodChatInstruction = "Given the following reflection: '%s', the implementation is working as expected, and the issues found are the result of problems in the tests?"
	goCodeBlockInstruction              = "Use a Go code block to write your response. For example:\n```go\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n```"

	testFunctionPattern = "(?s)```go(.*?)```"
)

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
	llm                     models.LLM
	goReflectionFewShotAdd  string
	goTestGenerationFewShot string
	goSelfReflectionFewShot string
	testFunctionRegex       *regexp.Regexp
}

func NewGoGenerator(llm models.LLM, promptsDir string) models.Generator {

	goReflectionFewShotAdd := readTxt(filepath.Join(promptsDir, "GoReflectionFewShotAdd.md"))
	goTestGenerationFewShot := readTxt(filepath.Join(promptsDir, "GoTestGenerationFewShot.md"))
	goSelfReflectionFewShot := readTxt(filepath.Join(promptsDir, "GoSelfReflectionFewShot.md"))

	return &goGenerator{
		llm:                     llm,
		goReflectionFewShotAdd:  goReflectionFewShotAdd,
		goTestGenerationFewShot: goTestGenerationFewShot,
		goSelfReflectionFewShot: goSelfReflectionFewShot,
		testFunctionRegex:       regexp.MustCompile(testFunctionPattern),
	}
}

func (g *goGenerator) GenerateCode(ctx context.Context, code string) (*string, error) {
	systemInstruction := fmt.Sprintf("%s\n%s", goSimpleChatInstruction, goCodeBlockInstruction)
	fmt.Println(systemInstruction)

	messages := []models.ChatMessage{
		{Type: models.SystemMessage, Content: systemInstruction},
		{Type: models.UserMessage, Content: code},
	}

	return g.llm.Chat(ctx, messages)
}

func (g *goGenerator) GenerateCodeWithReflection(ctx context.Context, code string, previousResult string, feedback string, selfReflection string) (*string, error) {
	systemInstruction := fmt.Sprintf("%s\n%s", goReflectionChatInstruction, goCodeBlockInstruction)
	fmt.Println(systemInstruction)

	messages := []models.ChatMessage{
		{Type: models.SystemMessage, Content: systemInstruction + "\n\n" + g.goReflectionFewShotAdd},
		{Type: models.UserMessage, Content: code},
		{Type: models.AssistantMessage, Content: fmt.Sprintf("```go\n%s\n```", previousResult)},
		{Type: models.UserMessage, Content: fmt.Sprintf("[unit test results from previous impl]:\n%s\n\n[reflection on previous impl]:", feedback)},
		{Type: models.AssistantMessage, Content: selfReflection},
		{Type: models.UserMessage, Content: fmt.Sprintf("Try to convert this code again:\n%s", code)},
	}

	return g.llm.Chat(ctx, messages)
}

func (g *goGenerator) GenerateSelfReflection(ctx context.Context, code string, feedback string) (*string, error) {
	systemInstruction := fmt.Sprintf("%s\n%s", goSimpleChatInstruction, goCodeBlockInstruction)
	fmt.Println(systemInstruction)

	messages := []models.ChatMessage{
		{Type: models.SystemMessage, Content: goSelfReflectionChatInstruction},
		{Type: models.UserMessage, Content: fmt.Sprintf("%s\n\n[function impl]:\n```go\n%s\n```\n\n[unit test results]:\n%s\n\n[self-reflection]:", g.goSelfReflectionFewShot, code, feedback)},
	}

	return g.llm.Chat(ctx, messages)
}

func (g *goGenerator) GenerateTests(ctx context.Context, code string, generatedCode string) ([]string, error) {
	messages := []models.ChatMessage{
		{Type: models.SystemMessage, Content: goTestGenerationChatInstruction},
		{Type: models.UserMessage, Content: fmt.Sprintf(goTestGenerationHumanInstruction, code, generatedCode)},
	}

	generatedTests, err := g.llm.Chat(ctx, messages)
	if err != nil {
		return nil, err
	}

	matches := g.testFunctionRegex.FindAllString(*generatedTests, -1)
	return matches, nil
}

func (g *goGenerator) QueryImplementationIsGood(ctx context.Context, reflection string) (*string, error) {
	messages := []models.ChatMessage{
		{Type: models.SystemMessage, Content: goYesNoChatInstruction},
		{Type: models.UserMessage, Content: fmt.Sprintf(goImplementationGoodChatInstruction, reflection)},
	}

	return g.llm.Chat(ctx, messages)
}
