package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
)

var invokeCount int
var myObjects []types.Object

const (
	EnvVarAzureFunctionPort = "FUNCTIONS_PORT"
)

func init() {
	// Load the SDK configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}

	// Initialize an S3 client
	svc := s3.NewFromConfig(cfg)

	// Define the bucket name as a variable so we can take its address
	bucketName := "examplebucket"
	input := &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	}

	// List objects in the bucket
	result, err := svc.ListObjectsV2(context.TODO(), input)
	if err != nil {
		log.Fatalf("Failed to list objects: %v", err)
	}
	myObjects = result.Contents
}

func LambdaHandler(ctx *gin.Context) {
	invokeCount++
	for i, obj := range myObjects {
		log.Printf("object[%d] size: %d key: %s", i, obj.Size, *obj.Key)
	}

	ctx.String(http.StatusOK, strconv.Itoa(invokeCount))
}

func main() {
	r := gin.Default()
	r.Handle(http.MethodPost, "/BookSaverHandler", LambdaHandler)

	port := os.Getenv(EnvVarAzureFunctionPort)
	if port == "" {
		port = "8080"
	}
	host := fmt.Sprintf("0.0.0.0:%s", port)
	fmt.Printf("Go server Listening...on port: %s\n", port)
	log.Fatal(r.Run(host))
}
