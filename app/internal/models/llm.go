package models

import (
	"github.com/tmc/langchaingo/prompts"
)

type LLM interface {
	Chat(messages []prompts.MessageFormatter, inputs map[string]any) (string, error)
}
