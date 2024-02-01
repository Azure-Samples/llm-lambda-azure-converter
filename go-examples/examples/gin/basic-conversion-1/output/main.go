package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	EnvVarAzureFunctionPort = "FUNCTIONS_PORT"
)

type MyEvent struct {
	Name string `json:"name"`
}

type MyResponse struct {
	Message string `json:"message"`
}

func HandleRequest(ctx *gin.Context) {
	if ctx.Request.Body == nil {
		errorMsg := "received nil event"
		ctx.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		return
	}

	var event MyEvent
	err := ctx.ShouldBindJSON(&event)
	if err != nil {
		errorMsg := fmt.Sprintf("error on reading request body: %v\n", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
		return
	}

	message := fmt.Sprintf("Hello %s!", event.Name)
	ctx.JSON(http.StatusOK, &MyResponse{Message: message})
}

func main() {
	r := gin.Default()
	r.Handle(http.MethodPost, "/HandleRequest", HandleRequest)

	r.Run()
}
