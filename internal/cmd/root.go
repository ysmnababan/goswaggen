package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// overridden at build time
// go build -ldflags="-X 'github.com/ysmnababan/goswaggen/internal/cmd.version=v0.2.3'" -o goswaggen.exe
var version = "v1.0.0"
var rootCmd = &cobra.Command{
	Use:     "goswaggen",
	Short:   "Goswaggen is a cli tool for generating handler comment block",
	Long:    "Goswaggen is a command-line tool that helps you automatically generate Swagger (OpenAPI) comment annotations for your Go HTTP handlers.",
	Version: version,
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
