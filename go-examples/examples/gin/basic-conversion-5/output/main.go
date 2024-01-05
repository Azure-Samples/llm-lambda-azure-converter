package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
)

const (
	EnvVarTableName         = "TABLE_NAME"
	EnvVarAzureFunctionPort = "FUNCTIONS_PORT"
)

var (
	dynamodbClient *dynamodb.Client
	tableName      string
)

func init() {
	var sdkConfig aws.Config = aws.Config{}
	dynamodbClient = dynamodb.NewFromConfig(sdkConfig)

	tableName = os.Getenv(EnvVarTableName)
}

type Book struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"description"`
}

func bookSaverHandler(ctx *gin.Context) {
	var book Book
	err := ctx.Bind(&book)
	if err != nil {
		errorMsg := fmt.Sprintf("error on reading request body: %v\n", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		return
	}

	item, err := attributevalue.MarshalMap(book)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = dynamodbClient.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String(tableName), Item: item,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("couldn't add item to table. Here's why: %v", err)})
		return

	}

	ctx.Status(http.StatusOK)
}

func main() {
	r := gin.Default()
	r.Handle(http.MethodPost, "/BookSaverHandler", bookSaverHandler)

	port := os.Getenv(EnvVarAzureFunctionPort)
	if port == "" {
		port = "8080"
	}
	host := fmt.Sprintf("0.0.0.0:%s", port)
	fmt.Printf("Go server Listening...on port: %s\n", port)
	log.Fatal(r.Run(host))
}
