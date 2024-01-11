For example:

Given the following code:

```go
type Request struct {
	InitCount int `json:"initCount"`
}

type Response struct {
	Count int `json:"count"`
}

func handler(ctx context.Context, request Request) (Response, error) {
	request.InitCount++
	return Response{Count: request.InitCount}, nil
}

func main() {
	lambda.Start(handler)
}
```

using the following endpoint:
/handle

the unit tests would be:
func TestHandler(t *testing.T) {
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

}
