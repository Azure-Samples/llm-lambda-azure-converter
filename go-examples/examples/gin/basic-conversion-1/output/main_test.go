package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleRequest(t *testing.T) {
	type args struct {
		request *http.Request
	}
	tests := []struct {
		name         string
		request      func() *http.Request
		expectedCode int
		expectedBody string
	}{
		{
			name: "success",
			request: func() *http.Request {
				req := `{"name":"Ana"}`
				return httptest.NewRequest(http.MethodPost, "/handle", strings.NewReader(req))
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"message":"Hello Ana!"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			router.POST("/handle", HandleRequest)

			// Create a ResponseRecorder
			w := httptest.NewRecorder()

			// Create an HTTP handler from the Gin router
			httpHandler := router

			// Serve the HTTP request with our ResponseRecorder
			httpHandler.ServeHTTP(w, tt.request())

			// Assert HTTP response status code
			assert.Equal(t, tt.expectedCode, w.Code)

			// Assert HTTP response body
			assert.Equal(t, tt.expectedBody, w.Body.String())

		})
	}
}
