package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/msft-latam-devsquad/lambda-to-azure-converter/go-examples/examples/storage"
)

type SaveRequest struct {
	Id string `json:"id" binding:"required"`
}

type Response struct {
	Message string `json:"message"`
}

func HandleRequest(c *gin.Context) {
	var req SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	azStore := storage.NewAzureStorage()
	err := azStore.Save(context.Background(), req.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	message := fmt.Sprintf("request %s was successfully saved", req.Id)
	c.JSON(http.StatusOK, &Response{Message: message})
}

func main() {
	r := gin.Default()
	r.POST("/save", HandleRequest)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
