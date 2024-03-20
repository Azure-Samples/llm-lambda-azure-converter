package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/MicahParks/go-aws-sam-lambda-example/util"
)

type lambdaTwoHandler struct {
	logger *log.Logger
	mux    sync.Mutex
	prev   string
}

type responseData struct {
	CustomPath     string `json:"customPath"`
	PrevCustomPath string `json:"prevCustomPath"`
}

// New creates a new handler for Lambda two.
func New(logger *log.Logger) lambda.Handler {
	return util.NewHandlerV1(&lambdaTwoHandler{
		logger: logger,
	})
}

// Handle implements util.LambdaHTTPV1 interface. It contains the logic for the handler.
func (handler *lambdaTwoHandler) Handle(_ context.Context, request *events.APIGatewayProxyRequest) (response *events.APIGatewayProxyResponse, err error) {
	response = &events.APIGatewayProxyResponse{}

	path, ok := request.PathParameters["customPath"]
	if !ok {
		handler.logger.Println("No custom path given, but AWS routed this request to this Lambda anyways.")
		path = "MISSING"
	}

	handler.mux.Lock()
	prev := handler.prev
	handler.prev = path // TODO Note in README.md that this won't work in AWS SAM, but will in AWS Lambda.
	handler.mux.Unlock()

	resp := responseData{
		CustomPath:     path,
		PrevCustomPath: prev,
	}

	data, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		handler.logger.Printf("Failed to JSON marshal response.\nError: %v", err)
		response.StatusCode = http.StatusInternalServerError
		return response, nil
	}

	response.StatusCode = http.StatusOK
	response.Body = string(data)

	return response, nil
}