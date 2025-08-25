package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ysmnababan/goswaggen/internal/generator"
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
		root := ""
		targetFunc := args[0]
		parser, err := parser.NewParser(root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while create parser: %v\n", err)
			os.Exit(1)
		}
		handlerReg, err := parser.ExtractFuncHandlerInfo(targetFunc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		gen := generator.NewGenerator(handlerReg)
		if shouldForce {
			
		} else {
			gen.PrintCommmentBlock()
		}
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
