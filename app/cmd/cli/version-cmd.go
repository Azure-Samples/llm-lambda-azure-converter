package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of AzFunc",
	Long:  `All software has versions. This is AzFunc's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("AzFunc v0.0.1")
	},
}