package lats

import (
	"strings"

	"github.com/spf13/viper"
)

const (
	AzureOpenAIApiKeyConfig         = "azure.openai.apikey"
	AzureOpenAIEndpointConfig       = "azure.openai.endpoint"
	AzureOpenAIApiVersionConfig     = "azure.openai.apiversion"
	AzureOpenAIDeploymentNameConfig = "azure.openai.deployment"
	ConverterMaxIterationsConfig    = "converter.maxIterations"
	ConverterMaxChildrenConfig      = "converter.maxChildren"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./..")
	viper.AddConfigPath("./../..")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

type LatsConfig struct {
	AzureOpenAIApiKey         string
	AzureOpenAIEndpoint       string
	AzureOpenAIApiVersion     string
	AzureOpenAIDeploymentName string
	ConverterMaxIterations    int
	ConverterMaxChildren      int
}

func NewLatsConfig(v viper.Viper) *LatsConfig {
	return &LatsConfig{
		AzureOpenAIApiKey:         v.GetString(AzureOpenAIApiKeyConfig),
		AzureOpenAIEndpoint:       v.GetString(AzureOpenAIEndpointConfig),
		AzureOpenAIApiVersion:     v.GetString(AzureOpenAIApiVersionConfig),
		AzureOpenAIDeploymentName: v.GetString(AzureOpenAIDeploymentNameConfig),
		ConverterMaxIterations:    v.GetInt(ConverterMaxIterationsConfig),
		ConverterMaxChildren:      v.GetInt(ConverterMaxChildrenConfig),
	}
}
