package models

import (
	"github.com/tmc/langchaingo/prompts"
)

type AIChat interface {
	Chat(messages []prompts.MessageFormatter) (string, error)
}
