package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gocli",
	Short: "gocli is a cli tool for testing cobra",
	Long:  "gocli is a cli tool for testing coba. This is a experiment for the goswaggen project",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error while executing gocli\n")
		os.Exit(1)
	}
}
