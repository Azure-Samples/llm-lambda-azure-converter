package models

import "context"

type Generator interface {
	GenerateCode(ctx context.Context, code string) (*string, error)
	GenerateCodeWithReflection(ctx context.Context, code string, previousResult string, feedback string, selfReflection string) (*string, error)
	GenerateTests(ctx context.Context, code string, generatedCode string) ([]string, error)
	GenerateSelfReflection(ctx context.Context, code string, feedback string) (*string, error)
	QueryFuncSignature(ctx context.Context, code string) (*string, error)
}
