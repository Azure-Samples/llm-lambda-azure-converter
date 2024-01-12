package models

import "context"

type Converter interface {
	Convert(ctx context.Context, code string, tests []string, generateTests bool) (*string, bool, error)
}
