package models

import (
	"context"
	"time"
)

type Converter interface {
	Convert(ctx context.Context, code string, tests []string, generateTests bool) (*ConverterResponse, error)
}

type ConverterResponse struct {
	Code            string
	Tests           []string
	TotalIterations int
	SelectedNode    string
	TotalTime       time.Duration
	Found           bool
}
