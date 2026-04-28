package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/ysmnababan/goswaggen/internal/config"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Aliases: []string{"i"},
	Short:   "Init config file",
	Long:    "Initialize a config file for customization",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		f, err := os.Create(config.YamlConfigName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while generating config file: %v\n", err)
			os.Exit(1)
		}
		err = InitConfig(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while generating config file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Config file generated...")
	},
}

func InitConfig(out io.Writer) error {
	w := bufio.NewWriter(out)
	_, err := w.WriteString(config.YamlConfigTemplate)
	if err != nil {
		return err
	}
	_ = w.Flush()
	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
