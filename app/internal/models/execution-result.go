package models

type ExecutionResult struct {
	IsPassing bool
	Feedback  string
	State     []bool
}
