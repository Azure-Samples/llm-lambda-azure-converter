package models

import (
	"context"
	"time"
)

type Converter interface {
	Convert(ctx context.Context, code string, tests []string, generateTests bool) (*string, *ConverterStatistics, error)
}

type ConverterStatistics struct {
	TotalIterations int
	SelectedNode    string
	TotalTime       time.Duration
	Found           bool
}
