package models

type Generator interface {
	GenerateCode(code string) (string, error)
	GenerateTests(code string) ([]string, error)
	GenerateReflection(code string) (string, error)
}