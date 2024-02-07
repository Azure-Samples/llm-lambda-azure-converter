package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	InitCount int `json:"initCount"`
}

type Response struct {
	Count int `json:"count"`
}

func handler(ctx context.Context, request Request) (Response, error) {
	request.InitCount++
	return Response{Count: request.InitCount}, nil
}

func main() {
	lambda.Start(handler)
}
