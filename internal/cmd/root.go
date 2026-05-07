package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// overridden at build time by GoReleaser
// go build -ldflags="-X 'github.com/ysmnababan/goswaggen/internal/cmd.version=v1.1.0'"
var version = "dev"

func getVersion() string {
	// if ldflags injected a real version, use it
	if version != "dev" {
		return version
	}
	// fallback: read from module build info (works with go install)
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

var rootCmd = &cobra.Command{
	Use:   "goswaggen",
	Short: "Goswaggen is a cli tool for generating handler comment block",
	Long:  "Goswaggen is a command-line tool that helps you automatically generate Swagger (OpenAPI) comment annotations for your Go HTTP handlers.",
	// Version: getVersion(),
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

func Execute() {
	rootCmd.Version = getVersion()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error while executing gocli\n")
		os.Exit(1)
	}
}
