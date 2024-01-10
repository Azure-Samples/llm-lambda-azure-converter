package models

type Executor interface {
	Execute(code string, tests []string) (*ExecutionResult, error)
}
