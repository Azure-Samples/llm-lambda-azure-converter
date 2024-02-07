For example:

Here is the Go code for the AWS Lambda function: 
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

Here is the Go code for the GinGonic http server:
```go
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
	log.Fatal(r.Run())
}
```

the unit tests would be:
```go
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
```
