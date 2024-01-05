package lats

import (
	"strings"

	"github.com/spf13/viper"
)

const (
	AzureOpenAIApiKeyConfig     = "azure.openai.apikey"
	AzureOpenAIEndpointConfig   = "azure.openai.endpoint"
	AzureOpenAIApiVersionConfig = "azure.openai.apiversion"
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
	AzureOpenAIApiKey     string
	AzureOpenAIEndpoint   string
	AzureOpenAIApiVersion string
}

func NewLatsConfig(v viper.Viper) *LatsConfig {
	return &LatsConfig{
		AzureOpenAIApiKey:     v.GetString(AzureOpenAIApiKeyConfig),
		AzureOpenAIEndpoint:   v.GetString(AzureOpenAIEndpointConfig),
		AzureOpenAIApiVersion: v.GetString(AzureOpenAIApiVersionConfig),
	}
}
