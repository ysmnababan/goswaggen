package cmd

import (
	"fmt"
	"os"

	"github.com/ysmnababan/goswaggen/internal/parser"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "controller lists",
	Long:    "list all the controller grouped by its package",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("All Controllers")
		prsr, err := parser.NewParser(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while create parser: %v\n", err)
			os.Exit(1)
		}
		c := prsr.GetAllHandlers()
		for p, funcs := range c {
			fmt.Println(p, ":")
			for _, f := range *funcs {
				fmt.Println("	", f)
			}
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
