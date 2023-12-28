package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/msft-latam-devsquad/lambda-to-azure-converter/examples/storage"
)

type SaveRequest struct {
	Id string `json:"id"`
}

type Response struct {
	Message string `json:"message"`
}

func HandleRequest(ctx context.Context, req *SaveRequest) (*Response, error) {
	if req == nil {
		return nil, fmt.Errorf("request can't be nil")
	}
	
	azStore := storage.NewAzureStorage()
	err := azStore.Save(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf("request %s was successfully saved", req.Id)
	return &Response{Message: message}, nil
}

func main() {
	lambda.Start(HandleRequest)
}