package lats

import (
	"context"
	"strings"
	"testing"

	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/models"
	"github.com/spf13/viper"
)

const code = `
package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/msft-latam-devsquad/lambda-to-azure-converter/go-examples/examples/storage"
)

type SaveRequest struct {
	Id string
}

type Response struct {
	Message string
}

func HandleRequest(ctx context.Context, req *SaveRequest) (*Response, error) {
	if req == nil {
		return nil, fmt.Errorf("request can't be nil")
	}
	
	azStore := storage.NewAzureStorage()
	err := azStore.Save(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf("request %s was successfully saved", req.Id)
	return &Response{Message: message}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
`

const generatedCode = `
package main

        import (
                "fmt"
                "net/http"
                "github.com/gin-gonic/gin"

                "github.com/msft-latam-devsquad/lambda-to-azure-converter/go-examples/examples/storage"
        )

        // Define the request and response structures.
        type SaveRequest struct {
                Id string
        }

        type Response struct {
                Message string
        }

        // SaveHandler handles the HTTP POST request to save data.
        func SaveHandler(c *gin.Context) {
                var req SaveRequest

                // Bind the incoming JSON payload to the SaveRequest struct.
                if err := c.ShouldBindJSON(&req); err != nil {
                        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
                        return
                }

                // Initialize AzureStorage.
                azStore := storage.NewAzureStorage()

                // Save the data using the AzureStorage instance.
                if err := azStore.Save(c, req.Id); err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                        return
                }

                // Send back a successful response.
                message := fmt.Sprintf("Request %s was successfully saved", req.Id)
                c.JSON(http.StatusOK, Response{Message: message})
        }

        func main() {
                // Set up Gin router.
                router := gin.Default()

                // Define the endpoint matching the original Lambda function.
                router.POST("/save", SaveHandler)

                // Start the server.
                router.Run() // Default port is 8080
        }
`

const wrongGeneratedCode = `
package main

        import (
                "fmt"
                "net/http"
                "github.com/gin-gonic/gin"

                "github.com/msft-latam-devsquad/lambda-to-azure-converter/go-examples/examples/storage"
        )

        // Define the request and response structures.
        type SaveRequest struct {
                Id string
        }

        type Response struct {
                Message string
        }

        // SaveHandler handles the HTTP POST request to save data.
        func SaveHandler(c *gin.Context) {
                var req SaveRequest

                // Bind the incoming JSON payload to the SaveRequest struct.
                if err := c.ShouldBindJSON(&req); err != nil {
                        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
                        return
                }

                // Initialize AzureStorage.
                azStore := storage.NewAzureStorage()

                // Save the data using the AzureStorage instance.
                if err := azStore.Save(c, req.Id); err != nil {
                        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                        return
                }

                // Send back a successful response.
                message := fmt.Sprintf("Request %s was successfully saved", req.Id)
                c.JSON(http.StatusOK, MyResponse{Message: message})
        }

        func main() {
                // Set up Gin router.
                router := gin.Default()

                // Define the endpoint matching the original Lambda function.
                router.POST("/save", SaveHandler)

                // Start the server.
                router.Run() // Default port is 8080
        }
`

const feedback = "cmd\\example\\main.go:41:24: undefined: MyResponse"

func generateGoGenerator() (models.Generator, error) {
	v := viper.GetViper()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./..")
	v.AddConfigPath("./../..")

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	config := NewLatsConfig(*v)
	llm, err := NewOpenAIChat(*config)
	if err != nil {
		return nil, err
	}
	gen := NewGoGenerator(llm, "./prompts")
	return gen, nil
}

func Test_goGenerator_GenerateCode(t *testing.T) {
	type args struct {
		code string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test GenerateCode",
			args: args{
				code: code,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := generateGoGenerator()
			if err != nil {
				t.Errorf("error creating goGenerator: %v", err)
				return
			}
			got, err := generator.GenerateCode(context.Background(), tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("goGenerator.GenerateCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(*got) == 0 {
				t.Errorf("goGenerator.GenerateCode() didn't generate a response")
			}
		})
	}
}

func Test_goGenerator_GenerateCodeWithReflection(t *testing.T) {
	type args struct {
		code           string
		previousResult string
		feedback       string
		selfReflection string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test GenerateCodeWithReflection",
			args: args{
				code:           code,
				previousResult: wrongGeneratedCode,
				feedback:       feedback,
				selfReflection: "The implementation is failing during compilation because the wrong struct type is being referenced in the `SaveHandler` function. Instead of using `Response`, which is the defined structure, `MyResponse` is mistakenly used when creating the JSON response, and `MyResponse` is not defined anywhere in the code. To fix this issue, the correct type `Response` should be used to create the JSON response.",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := generateGoGenerator()
			if err != nil {
				t.Errorf("error creating goGenerator: %v", err)
				return
			}
			got, err := generator.GenerateCodeWithReflection(context.Background(), tt.args.code, tt.args.previousResult, tt.args.feedback, tt.args.selfReflection)
			if (err != nil) != tt.wantErr {
				t.Errorf("goGenerator.GenerateCodeWithReflection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(*got) == 0 {
				t.Errorf("goGenerator.GenerateCodeWithReflection() didn't generate a response")
			}
		})
	}
}

func Test_goGenerator_GenerateSelfReflection(t *testing.T) {
	type args struct {
		code     string
		feedback string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test GenerateSelfReflection",
			args: args{
				code:     wrongGeneratedCode,
				feedback: feedback,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := generateGoGenerator()
			if err != nil {
				t.Errorf("error creating goGenerator: %v", err)
				return
			}
			got, err := generator.GenerateSelfReflection(context.Background(), tt.args.code, tt.args.feedback)
			if (err != nil) != tt.wantErr {
				t.Errorf("goGenerator.GenerateSelfReflection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(*got) == 0 {
				t.Errorf("goGenerator.GenerateSelfReflection() didn't generate a response")
			}
		})
	}
}

func Test_goGenerator_GenerateTests(t *testing.T) {
	type args struct {
		code          string
		generatedCode string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test GenerateTests",
			args: args{
				code:          code,
				generatedCode: generatedCode,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := generateGoGenerator()
			if err != nil {
				t.Errorf("error creating goGenerator: %v", err)
				return
			}
			got, err := generator.GenerateTests(context.Background(), tt.args.code, tt.args.generatedCode)
			if (err != nil) != tt.wantErr {
				t.Errorf("goGenerator.GenerateTests() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) == 0 {
				t.Errorf("goGenerator.GenerateTests() didn't generate any response")
			}
		})
	}
}

func Test_goGenerator_QueryFuncSignature(t *testing.T) {
	type args struct {
		code string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test QueryFuncSignature",
			args: args{
				code: generatedCode,
			},
			want:    "func SaveHandler(c *gin.Context)",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := generateGoGenerator()
			if err != nil {
				t.Errorf("error creating goGenerator: %v", err)
				return
			}
			got, err := generator.QueryImplementationIsGood(context.Background(), tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("goGenerator.QueryFuncSignature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.Contains(*got, tt.want) {
				t.Errorf("goGenerator.QueryFuncSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}
