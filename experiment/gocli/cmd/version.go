package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Show the cli version information",
	Long:    "Show the cli version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version %s", Version())
	},
}

func Version() string {
	return "1.0.0"
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
