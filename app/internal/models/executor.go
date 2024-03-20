package models

type ExecutorOptions struct {
	ProjectPath   string
	TargetPath    string
	Filename      string
	CreateProject bool
}

type ExecutorOption func(*ExecutorOptions)

func WithExProjectPath(projectPath string) ExecutorOption {
	return func(o *ExecutorOptions) {
		if projectPath != "" {
			o.ProjectPath = projectPath
		}
	}
}

func WithExExTargetPath(targetPath string) ExecutorOption {
	return func(o *ExecutorOptions) {
		if targetPath != "" {
			o.TargetPath = targetPath
		}
	}
}

func WithExFilename(filename string) ExecutorOption {
	return func(o *ExecutorOptions) {
		if filename != "" {
			o.Filename = filename
		}
	}
}

func WithExCreateProject(createProject bool) ExecutorOption {
	return func(o *ExecutorOptions) {
		o.CreateProject = createProject
	}
}

type Executor interface {
	Execute(code string, tests []string, options ...ExecutorOption) (*ExecutionResult, error)
}
