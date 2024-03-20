package models

import (
	"context"
	"time"
)

type ConverterOptions struct {
	GenerateTests bool
	CreateProject bool
	ProjectPath   string
	TargetPath    string
	MainFile      string
}

type ConverterOption func(*ConverterOptions)

func WithGenerateTests(generateTests bool) ConverterOption {
	return func(o *ConverterOptions) {
		o.GenerateTests = generateTests
	}
}

func WithCreateProject(createProject bool) ConverterOption {
	return func(o *ConverterOptions) {
		o.CreateProject = createProject
	}
}

func WithProjectPath(projectPath string) ConverterOption {
	return func(o *ConverterOptions) {
		if projectPath != "" {
			o.ProjectPath = projectPath
		}
	}
}

func WithTargetPath(targetPath string) ConverterOption {
	return func(o *ConverterOptions) {
		if targetPath != "" {
			o.TargetPath = targetPath
		}
	}
}

func WithMainFile(mainFile string) ConverterOption {
	return func(o *ConverterOptions) {
		if mainFile != "" {
			o.MainFile = mainFile
		}
	}
}

type Converter interface {
	Convert(ctx context.Context, code string, tests []string, options ...ConverterOption) (*ConverterResponse, error)
}

type ConverterResponse struct {
	Code            string
	Tests           []string
	TotalIterations int
	TotalAttempts   int
	SelectedNode    string
	TotalTime       time.Duration
	Found           bool
}
