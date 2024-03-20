package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/msft-latam-devsquad/lambda-to-azure-converter/go-examples/examples/gin/complex-conversion-2/input/handler"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Llongfile)
	logger.Println("Lambda two has started.")

	// The main goroutine in a Lambda might never run its deferred statements.
	// This is because of how the Lambda is shutdown.
	// https://docs.aws.amazon.com/lambda/latest/dg/runtimes-context.html#runtimes-lifecycle-shutdown
	defer logger.Println("Lambda two has stopped.")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := handler.New(logger)

	lambda.StartWithOptions(h, lambda.WithContext(ctx))
}