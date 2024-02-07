package lats

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/models"
)

type openAIChat struct {
	config LatsConfig
	client *azopenai.Client
}

func NewOpenAIChat(config LatsConfig) (models.LLM, error) {
	keyCredential := azcore.NewKeyCredential(config.AzureOpenAIApiKey)
	client, err := azopenai.NewClientWithKeyCredential(config.AzureOpenAIEndpoint, keyCredential, nil)
	if err != nil {
		return nil, err
	}

	return &openAIChat{
		config: config,
		client: client,
	}, nil
}

func (o *openAIChat) Chat(ctx context.Context, messages []models.ChatMessage) (*string, error) {
	chatMessages := make([]azopenai.ChatRequestMessageClassification, 0)
	for _, message := range messages {
		switch message.Type {
		case models.SystemMessage:
			chatMessages = append(chatMessages, &azopenai.ChatRequestSystemMessage{Content: to.Ptr(message.Content)})
		case models.UserMessage:
			chatMessages = append(chatMessages, &azopenai.ChatRequestUserMessage{Content: azopenai.NewChatRequestUserMessageContent(message.Content)})
		case models.AssistantMessage:
			chatMessages = append(chatMessages, &azopenai.ChatRequestAssistantMessage{Content: to.Ptr(message.Content)})
		}
	}

	resp, err := o.client.GetChatCompletions(ctx, azopenai.ChatCompletionsOptions{
		Messages:       chatMessages,
		DeploymentName: &o.config.AzureOpenAIDeploymentName,
	}, nil)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, errors.New("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}
