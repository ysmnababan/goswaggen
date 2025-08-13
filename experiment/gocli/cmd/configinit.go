package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"i"},
	Short:   "init config file",
	Long:    "Initialize a config file for costumization",
	Run: func(cmd *cobra.Command, args []string) {
		err := InitConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while generating config file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Config file generated...")
	},
}

func InitConfig() error {
	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
