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

type Event struct {
	Name string `json:"name"`
}

func HandleRequest(ctx *gin.Context) {
	var event Event
	err := ctx.ShouldBindJSON(&event)
	if err != nil {
		errorMsg := fmt.Sprintf("error on reading request body: %v\n", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		return
	}

	message := fmt.Sprintf("Hello %s!", event.Name)
	ctx.String(http.StatusOK, message)
}

func main() {
	r := gin.Default()
	r.Handle(http.MethodPost, "/HandleRequest", HandleRequest)

	port := os.Getenv(EnvVarAzureFunctionPort)
	if port == "" {
		port = "8080"
	}
	host := fmt.Sprintf("0.0.0.0:%s", port)
	fmt.Printf("Go server Listening...on port: %s\n", port)
	log.Fatal(r.Run(host))
}
