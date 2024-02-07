package models

import "context"

type MessageType string

const (
	SystemMessage    MessageType = "system"
	UserMessage      MessageType = "user"
	AssistantMessage MessageType = "assistant"
)

type ChatMessage struct {
	Type    MessageType
	Content string
}

type LLM interface {
	Chat(ctx context.Context, messages []ChatMessage) (*string, error)
}
