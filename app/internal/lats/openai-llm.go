package lats

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/models"

)

type openAIChat struct {
	config LatsConfig
}

func NewOpenAIChat(config LatsConfig) models.LLM {
	return &openAIChat{
		config: config,
	}
}

func (o *openAIChat) Chat(messages []prompts.MessageFormatter, inputs map[string]any) (string, error) {
	prompt := prompts.NewChatPromptTemplate(messages)

	fullPrompt, err := prompt.Format(inputs)
	if err != nil {
		return "", err
	}

	llm, err := openai.New(
		openai.WithBaseURL(o.config.AzureOpenAIEndpoint), 
		openai.WithToken(o.config.AzureOpenAIApiKey),
		openai.WithAPIVersion(o.config.AzureOpenAIApiVersion),
	)
	if err != nil {
		return "", err
	}

	return llm.Call(context.Background(), fullPrompt, llms.WithTemperature(0))
}
