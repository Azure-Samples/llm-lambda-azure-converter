package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

const (
	EnvVarTableName = "TABLE_NAME"
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

func bookHandler(ctx context.Context, book Book) error {
	item, err := attributevalue.MarshalMap(book)
	if err != nil {
		return err
	}
	_, err = dynamodbClient.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String(tableName), Item: item,
	})
	if err != nil {
		return fmt.Errorf("couldn't add item to table. Here's why: %v", err)
	}
	return err
}

func main() {
	lambda.Start(bookHandler)
}
