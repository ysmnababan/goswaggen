package parser

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/ysmnababan/goswaggen/internal/parser/context"
	"github.com/ysmnababan/goswaggen/internal/parser/tracking"
	"golang.org/x/tools/go/packages"
)

var FSET *token.FileSet
var MAIN_PACKAGE_NAME = "main"

type parser struct {
	fset         *token.FileSet
	root         string
	pkgs         []*packages.Package
	mainFuncDecl *ast.FuncDecl
}

func NewParser(root string) (*parser, error) {
	if root == "" {
		return nil, fmt.Errorf("root can't be empty")
	}

	fset := token.NewFileSet()
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			// packages.NeedCompiledGoFiles |
			packages.NeedImports |
			// packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
		Dir:  root, // relative to where you run `go run`
		Fset: fset,
	}

	pkgs, err := packages.Load(cfg, "./...") // load add the package
	if err != nil {
		return nil, fmt.Errorf("error loading packages: %w", err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no package found")
	}
	mainFuncDecl, _ := searchDeclFun(pkgs, "main", &MAIN_PACKAGE_NAME)
	if mainFuncDecl == nil {
		return nil, fmt.Errorf("no main file found")
	}
	return &parser{
		fset:         fset,
		root:         root,
		pkgs:         pkgs,
		mainFuncDecl: mainFuncDecl,
	}, nil
}

func (p *parser) GetAllHandlers() []string {
	ctx := context.NewRegistrationContext(p.pkgs, p.mainFuncDecl)
	handlerRegs := tracking.FindHandlerRegistration(ctx)
	if len(handlerRegs) == 0 {
		return nil
	}
	out := make([]string, 0, len(handlerRegs))

	for _, h := range handlerRegs {
		out = append(out, h.GetFuncNameWithPackage())
	}
	return out
}
