package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "goswaggen",
	Short: "Goswaggen is a cli tool for generating handler comment block",
	Long:  "Goswaggen is a command-line tool that helps you automatically generate Swagger (OpenAPI) comment annotations for your Go HTTP handlers.",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error while executing gocli\n")
		os.Exit(1)
	}
}
