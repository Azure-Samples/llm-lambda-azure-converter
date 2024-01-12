package lats

import (
	"github.com/spf13/viper"
)

const (
	AzureOpenAIApiKeyConfig         = "azure.openai.apikey"
	AzureOpenAIEndpointConfig       = "azure.openai.endpoint"
	AzureOpenAIApiVersionConfig     = "azure.openai.apiversion"
	AzureOpenAIDeploymentNameConfig = "azure.openai.deployment"
	ConverterMaxIterationsConfig    = "converter.maxIterations"
	ConverterMaxChildrenConfig      = "converter.maxChildren"
	ServerPortConfig                = "server.port"
)

type LatsConfig struct {
	AzureOpenAIApiKey         string
	AzureOpenAIEndpoint       string
	AzureOpenAIApiVersion     string
	AzureOpenAIDeploymentName string
	ConverterMaxIterations    int
	ConverterMaxChildren      int
	ServerPort                int
}

func NewLatsConfig(v viper.Viper) *LatsConfig {
	return &LatsConfig{
		AzureOpenAIApiKey:         v.GetString(AzureOpenAIApiKeyConfig),
		AzureOpenAIEndpoint:       v.GetString(AzureOpenAIEndpointConfig),
		AzureOpenAIApiVersion:     v.GetString(AzureOpenAIApiVersionConfig),
		AzureOpenAIDeploymentName: v.GetString(AzureOpenAIDeploymentNameConfig),
		ConverterMaxIterations:    v.GetInt(ConverterMaxIterationsConfig),
		ConverterMaxChildren:      v.GetInt(ConverterMaxChildrenConfig),
		ServerPort:                v.GetInt(ServerPortConfig),
	}
}
