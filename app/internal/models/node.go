package models

type Node struct {
	Code           string
	Feedback       string
	SelfReflection string
	Score          float32

	ParentNode *Node
	ChildNodes []*Node
}
