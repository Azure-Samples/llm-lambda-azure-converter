package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/lats"
	"github.com/msft-latam-devsquad/lambda-to-azure-converter/cli/internal/models"
	"github.com/spf13/viper"
)

type ConversionLanguage string

const (
	ConversionLanguageGo ConversionLanguage = "go"
)

const (
	promptsDir = "../../internal/lats/prompts"
)

type ConversionRequest struct {
	Language string   `json:"language"`
	Code     string   `json:"code"`
	Tests    []string `json:"tests"`
}

type ConversionResponse struct {
	Code string `json:"code"`
	Pass bool   `json:"pass"`
}

type Server interface {
	Run() error
}

type server struct {
	converterMap map[string]models.Converter
	port         int
}

func NewServer() (Server, error) {
	v, err := configViper()
	if err != nil {
		return nil, fmt.Errorf("error loading the config file: %v", err)
	}
	config := lats.NewLatsConfig(*v)
	llm, err := lats.NewOpenAIChat(*config)
	if err != nil {
		return nil, fmt.Errorf("error creating the LLM: %v", err)
	}

	goExecutor := lats.NewGoExecutor()
	goGenerator := lats.NewGoGenerator(llm, promptsDir)
	goConverter := lats.NewConverter(goGenerator, goExecutor, *config)

	return &server{
		converterMap: map[string]models.Converter{
			string(ConversionLanguageGo): goConverter,
		},
		port: config.ServerPort,
	}, nil
}

func (s *server) convertHandler(c *gin.Context) {
	// TODO: Make it async
	var request ConversionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("there was an error reading the request: %s", err.Error())})
		return
	}

	var converter models.Converter
	found := false
	for lang := range s.converterMap {
		if lang == request.Language {
			found = true
			converter = s.converterMap[lang]
		}
	}

	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported language: %s", request.Language)})
		return
	}

	code, pass, err := converter.Convert(c.Request.Context(), request.Code, request.Tests, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("there was an error doing the conversion: %s", err.Error())})
		return
	}

	response := ConversionResponse{
		Code: *code,
		Pass: pass,
	}

	c.JSON(http.StatusOK, response)
}

func (s *server) Run() error {
	r := gin.Default()
	r.Handle(http.MethodPost, "/convert", s.convertHandler)

	if s.port == 0 {
		s.port = 8080
	}
	host := fmt.Sprintf("0.0.0.0:%d", s.port)
	fmt.Printf("Go server Listening...on port: %d\n", s.port)

	return r.Run(host)
}

func configViper() (*viper.Viper, error) {
	v := viper.GetViper()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./..")
	v.AddConfigPath("./../..")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	return v, nil
}