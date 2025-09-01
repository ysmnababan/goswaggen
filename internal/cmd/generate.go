package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/ysmnababan/goswaggen/internal/generator"
	"github.com/ysmnababan/goswaggen/internal/injector"
	"github.com/ysmnababan/goswaggen/internal/model"
	"github.com/ysmnababan/goswaggen/internal/parser"
)

var shouldForce bool

type GeneratePayload struct {
	root        string
	targetFunc  string
	srcFile     io.Writer
	shouldForce bool
	config      model.Config
}

var generateCmd = &cobra.Command{
	Use:     "generate [handler to annotate]",
	Aliases: []string{"g", "gen"},
	Short:   "Generate Swagger comment block",
	Long:    "Generate Swagger comment block",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		root, _ := os.Getwd()
		targetFunc := args[0]
		payload := GeneratePayload{
			root:        root,
			targetFunc:  targetFunc,
			srcFile:     os.Stdout,
			shouldForce: shouldForce,
			config:      model.Cfg,
		}
		err := Generate(payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while generating comment block: %v\n", err)
			os.Exit(1)
		}
	},
}

func Generate(payload GeneratePayload) error {
	parser, err := parser.NewParser(payload.root, &payload.config)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "error while create parser: %v\n", err)
		// os.Exit(1)
		return err
	}
	handlerReg, err := parser.ExtractFuncHandlerInfo(payload.targetFunc)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "%v\n", err)
		// os.Exit(1)
		return err
	}

	gen := generator.NewGenerator(handlerReg, &payload.config)
	cmt := gen.CreateCommentBlock()
	if shouldForce {
		// TODO: Check the fset
		inject := injector.NewInjector(handlerReg.Pkg.Fset, handlerReg.File, handlerReg.FuncDecl)
		err := inject.InjectComment(cmt, payload.srcFile)
		if err != nil {
			// fmt.Fprintf(os.Stderr, "%v\n", err)
			// os.Exit(1)
			return err
		}
	} else {
		fmt.Print(`Copy this swagger comment to your code:
// Swaggo comment block

`)
		gen.PrintCommmentBlock()
	}
	return nil
}

func init() {
	generateCmd.Flags().BoolVarP(&shouldForce, "force", "f", false, "update the comment block directly on the source file")
	rootCmd.AddCommand(generateCmd)
}
