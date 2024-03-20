// Taken from https://www.golinuxcloud.com/golang-aws-lambda/

package main

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Person struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func list() (events.APIGatewayProxyResponse, error) {
	people := []Person{
		{Id: 1, Name: "GoLinuxCloud", Email: "golinu@gmail.com"},
		{Id: 2, Name: "Admin", Email: "admin@gmail.com"},
	}

	res, _ := json.Marshal(&people)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(res),
	}, nil

}

func main() {
	lambda.Start(list)
}
