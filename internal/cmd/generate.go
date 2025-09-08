package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ysmnababan/goswaggen/internal/config"
	"github.com/ysmnababan/goswaggen/internal/generator"
	"github.com/ysmnababan/goswaggen/internal/injector"
	"github.com/ysmnababan/goswaggen/internal/parser"
)

var shouldForce bool

type GeneratePayload struct {
	root        string
	targetFunc  string
	srcFile     io.Writer
	shouldForce bool
	config      *config.Config
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
			shouldForce: shouldForce,
			config:      config.Cfg,
		}
		if !shouldForce {
			payload.srcFile = os.Stdout
		}
		err := Generate(payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while generating comment block: %v\n", err)
			os.Exit(1)
		}
	},
}

func Generate(payload GeneratePayload) error {
	parser, err := parser.NewParser(payload.root, payload.config)
	if err != nil {
		return err
	}
	handlerReg, err := parser.ExtractFuncHandlerInfo(payload.targetFunc)
	if err != nil {
		return err
	}

	gen := generator.NewGenerator(handlerReg, payload.config)
	cmt := gen.CreateCommentBlock()
	if payload.shouldForce {
		inject := injector.NewInjector(handlerReg.Pkg.Fset, handlerReg.File, handlerReg.FuncDecl)
		fileToken := handlerReg.Pkg.Fset.File(handlerReg.File.Pos())
		f, err := os.Create(fileToken.Name())
		if err != nil {
			return err
		}
		defer func() {
			_ = f.Close()
		}()
		err = inject.InjectComment(cmt, f)
		if err != nil {
			return err
		}
	} else {
		fmt.Fprintf(payload.srcFile,
			"Copy this swagger comment to your code: \n\n%v",
			strings.Join(cmt, "\n"),
		)
	}
	return nil
}

func init() {
	generateCmd.Flags().BoolVarP(&shouldForce, "force", "f", false, "update the comment block directly on the source file")
	rootCmd.AddCommand(generateCmd)
}
