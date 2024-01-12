package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	userLicense string

	rootCmd = &cobra.Command{
		Use:   "azfunc",
		Short: "AzFunc is an app to convert lambda functions to azure functions ",
		Long: `AzFunc is an app powered by Azure OpenAI that converts lambda functions to azure functions.
				  it is using LATS for its underlying implementation.
				  Complete documentation is available at https://github.com/msft-latam-devsquad/azopenai-lambda-azfn-converter.git`,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/config.yaml)")
	viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./..")
		viper.AddConfigPath("./../..")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))		
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file: ", viper.ConfigFileUsed())
	}
}