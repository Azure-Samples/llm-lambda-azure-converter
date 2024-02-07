package lats

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

const configText = `
azure:
  openai:
    apikey: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    endpoint: "https://lats.openai.azure.com/"
    apiversion: "2023-12-01"
    deployment: "GPT-4"
converter:
  maxIterations: 5
  maxChildren: 3
server:
  port: 8080
`

func TestNewLatsConfig(t *testing.T) {
	type args struct {
		v func() viper.Viper
	}
	tests := []struct {
		name    string
		args    args
		want    *LatsConfig
		wantErr bool
	}{
		{
			name: "Test NewLatsConfig",
			args: args{
				v: func() viper.Viper {
					v := viper.New()
					v.SetConfigType("yaml")
					v.ReadConfig(strings.NewReader(configText))
					return *v
				},
			},
			want: &LatsConfig{
				AzureOpenAIApiKey:         "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				AzureOpenAIEndpoint:       "https://lats.openai.azure.com/",
				AzureOpenAIApiVersion:     "2023-12-01",
				AzureOpenAIDeploymentName: "GPT-4",
				ConverterMaxIterations:    5,
				ConverterMaxChildren:      3,
				ServerPort:                8080,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.args.v()
			got := NewLatsConfig(v)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLatsConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
