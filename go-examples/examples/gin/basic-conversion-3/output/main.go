package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	EnvVarAzureFunctionPort = "FUNCTIONS_PORT"
)

type Request struct {
	InitCount int `json:"initCount"`
}

type Response struct {
	Count int `json:"count"`
}

func handler(ctx *gin.Context) {
	var request Request
	err := ctx.Bind(&request)
	if err != nil {
		errorMsg := fmt.Sprintf("error on reading request body: %v\n", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		return
	}

	request.InitCount++
	ctx.JSON(http.StatusOK, &Response{Count: request.InitCount})
}

func main() {
	r := gin.Default()
	r.Handle(http.MethodPost, "/handler", handler)

	port := os.Getenv(EnvVarAzureFunctionPort)
	if port == "" {
		port = "8080"
	}
	host := fmt.Sprintf("0.0.0.0:%s", port)
	fmt.Printf("Go server Listening...on port: %s\n", port)
	log.Fatal(r.Run(host))
}
