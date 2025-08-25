package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ysmnababan/goswaggen/internal/parser"
)

var shouldForce bool
var generateCmd = &cobra.Command{
	Use:     "generate [handler to annotate]",
	Aliases: []string{"g", "gen"},
	Short:   "Generate Swagger comment block",
	Long:    "Generate Swagger comment block",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		parser, err := parser.NewParser(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while create parser: %v\n", err)
			os.Exit(1)
		}
		var handler IHandler
		handler, err = parser.ExtractFuncHandlerInfo(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		_ = handler
	},
}

func Generate(targetFunc string, shouldForce bool) error {
	if shouldForce {
		fmt.Println("Swagger comment updated successfuly")
	} else {
		fmt.Println(`Copy this swagger comment to your code:
// Swaggo comment block`)
	}
	return nil
}

func init() {
	generateCmd.Flags().BoolVarP(&shouldForce, "force", "f", false, "update the comment block directly on the source file")
	rootCmd.AddCommand(generateCmd)
}
