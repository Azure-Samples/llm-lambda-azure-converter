package models

type ExecutionResult struct {
	IsPassing bool
	Feedback  string
	Score     float32
}
