package lats

import (
	"context"
	"testing"

	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/models"
	"github.com/spf13/viper"
)

const codeToConvert = `
type MyEvent struct {
	Name string ` + "`json:\"name\"`" + `
}

type MyResponse struct {
	Message string ` + "`json:\"message\"`" + `
}

func HandleRequest(ctx context.Context, event *MyEvent) (*MyResponse, error) {
	if event == nil {
		return nil, fmt.Errorf("received nil event")
	}
	message := fmt.Sprintf("Hello %s!", event.Name)
	return &MyResponse{Message: message}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
`

const test1 = "```go\n" + `
package main

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
				req := "{\"name\":\"Ana\"}"
				return httptest.NewRequest(http.MethodPost, "/handle", strings.NewReader(req))
			},
			expectedCode: http.StatusOK,
			expectedBody: "{\"message\":\"Hello Ana!\"}",
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
}` + "\n```"

const test2 = "```go\n" + `
package main

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
				req := "{\"name\":\"Cassien\"}"
				return httptest.NewRequest(http.MethodPost, "/handle", strings.NewReader(req))
			},
			expectedCode: http.StatusOK,
			expectedBody: "{\"message\":\"Hello Cassien!\"}",
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
}` + "\n```"

const test3 = "```go\n" + `
package main

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
				req := "{\"name\":\"Hazel\"}"
				return httptest.NewRequest(http.MethodPost, "/handle", strings.NewReader(req))
			},
			expectedCode: http.StatusOK,
			expectedBody: "{\"message\":\"Hello Hazel!\"}",
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
}` + "\n```"

func Test_converter_Convert(t *testing.T) {
	type args struct {
		code          string
		originalTests []string
		generateTests bool
	}
	tests := []struct {
		name         string
		getConverter func(t *testing.T) models.Converter
		args         args
		wantPass     bool
		wantErr      bool
	}{
		{
			name: "Go Converter",
			getConverter: func(t *testing.T) models.Converter {
				executor := NewGoExecutor()
				v := viper.GetViper()
				v.SetConfigName("config")
				v.SetConfigType("yaml")
				v.AddConfigPath(".")
				v.AddConfigPath("./..")
				v.AddConfigPath("./../..")
			
				err := v.ReadInConfig()
				if err != nil {
					t.Errorf("error reading config file: %v", err)
				}
				config := NewLatsConfig(*v)
				llm, err := NewOpenAIChat(*config)
				if err != nil {
					t.Errorf("error creating the LLM: %v", err)
				}
				generator := NewGoGenerator(llm, "/prompts")

				return NewConverter(generator, executor, *config)
			},
			args: args{
				code:          codeToConvert,
				originalTests: []string{test1, test2, test3},
				generateTests: true,
			},
			wantPass: true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := tt.getConverter(t)
			got, statistics, err := converter.Convert(context.Background(), tt.args.code, tt.args.originalTests, tt.args.generateTests)
			if (err != nil) != tt.wantErr {
				t.Errorf("converter.Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if statistics.Found != tt.wantPass {
				t.Errorf("converter.Convert() pass = %v, want %v", statistics, tt.wantPass)
			}
			if len(*got) == 0 {
				t.Errorf("converter.Convert() didn't return any solution got = %v", got)
			}
		})
	}
}
