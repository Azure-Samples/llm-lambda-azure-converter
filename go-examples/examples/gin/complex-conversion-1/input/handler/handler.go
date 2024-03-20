package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tidwall/gjson"

	"github.com/MicahParks/go-aws-sam-lambda-example/util"
)

// ErrNoPokemon indicates that the Pokemon API failed in some way
var ErrNoPokemon = errors.New("failed to get Pokemon name")

type lambdaOneHandler struct {
	customString string
	logger       *log.Logger
}

type responseData struct {
	CustomString  string    `json:"customString"`
	RandomPokemon string    `json:"randomPokemon"`
	SourceIP      string    `json:"sourceIP"`
	Time          time.Time `json:"time"`
	UserAgent     string    `json:"userAgent"`
}

// New creates a new handler for Lambda one.
func New(logger *log.Logger, customString string) lambda.Handler {
	return util.NewHandlerV1(lambdaOneHandler{
		customString: customString,
		logger:       logger,
	})
}

// Handle implements util.LambdaHTTPV1 interface. It contains the logic for the handler.
func (handler lambdaOneHandler) Handle(ctx context.Context, request *events.APIGatewayProxyRequest) (response *events.APIGatewayProxyResponse, err error) {
	response = &events.APIGatewayProxyResponse{}

	now := time.Now()
	sourceIP := request.RequestContext.Identity.SourceIP
	userAgent := request.RequestContext.Identity.UserAgent

	pokemon, err := randomPokemon(ctx)
	if err != nil {
		const errMsg = "Failed to get random pokemon."
		if errors.Is(err, ErrNoPokemon) {
			handler.logger.Printf("%s The gjson path syntax is probably wrong.\nError: %s", errMsg, err.Error())
			pokemon = "garbodor API"
		} else {
			handler.logger.Printf("%s\nError: %s", errMsg, err.Error())
			pokemon = "trubbish API"
		}
	}

	resp := responseData{
		CustomString:  handler.customString,
		RandomPokemon: pokemon,
		SourceIP:      sourceIP,
		Time:          now,
		UserAgent:     userAgent,
	}

	data, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		handler.logger.Printf("Failed to JSON marshal response.\nError: %v", err)
		response.StatusCode = http.StatusInternalServerError
		return response, nil
	}

	response.StatusCode = http.StatusOK
	response.Body = string(data)

	return response, nil
}

func randomPokemon(ctx context.Context) (pokemon string, err error) {
	const apiURL = "https://pokeapi.co/api/v2/pokemon/"
	const errMsg = "Pokemon API"

	// There are 898 Pokemon in this API.
	i := rand.Int63n(898)

	u := apiURL + strconv.FormatInt(i, 10)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, bytes.NewReader(nil))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %s: %w", errMsg, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform request: %s: %w", errMsg, err)
	}

	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %s: %w", errMsg, err)
	}

	pokemon = gjson.Get(string(body), "species.name").Str
	if pokemon == "" {
		return "", fmt.Errorf("the Pokemon API response did not contain the species name: %w", ErrNoPokemon)
	}

	return pokemon, nil
}