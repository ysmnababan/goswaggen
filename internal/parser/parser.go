package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/ysmnababan/goswaggen/internal/model"
	"github.com/ysmnababan/goswaggen/internal/parser/context"
	"github.com/ysmnababan/goswaggen/internal/parser/inspector"
	"github.com/ysmnababan/goswaggen/internal/parser/inspector/payloadinspector"
	"github.com/ysmnababan/goswaggen/internal/parser/inspector/returninspector"
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
	ctx          *context.RegistrationContext
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
		ctx:          context.NewRegistrationContext(pkgs, mainFuncDecl),
	}, nil
}

func (p *parser) GetAllHandlers() map[string]*[]string {
	ctx := context.NewRegistrationContext(p.pkgs, p.mainFuncDecl)
	handlerRegs := tracking.FindHandlerRegistration(ctx)
	if len(handlerRegs) == 0 {
		return nil
	}
	out := make(map[string]*[]string)

	for _, h := range handlerRegs {
		funcNames, ok := out[h.GetPackageName()]
		if !ok {
			out[h.GetPackageName()] = &[]string{h.GetFuncName()}
		} else {
			*funcNames = append(*funcNames, h.GetFuncName())
		}
	}
	return out
}

// GetHandlerByFuncName
// Returns all matching handlers registration by name.
// The func name can be the name only or combination of name and package name.
// e.g. : name = `Login` or `auth.Login`.
func (p *parser) getHandlerByFuncName(name string) (*model.HandlerRegistration, error) {
	out := []*model.HandlerRegistration{}
	handlerRegs := tracking.FindHandlerRegistration(p.ctx)

	if strings.Contains(name, ".") {
		for _, h := range handlerRegs {
			if h.GetFuncNameWithPackage() == name {
				out = append(out, h)
			}
		}
	} else {
		for _, h := range handlerRegs {
			if h.GetFuncName() == name {
				out = append(out, h)
			}
		}
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no handler found ... ")
	}

	if len(out) != 1 {
		handlers := ""
		for _, h := range out {
			handlers += fmt.Sprintf("    %s\n", h.GetFuncNameWithPackage())
		}
		return nil, fmt.Errorf("multiple handlers found.\n%s", handlers)
	}
	return out[0], nil
}

func (p *parser) ExtractFuncHandlerInfo(name string) (*model.HandlerRegistration, error) {
	handlerFunc, err := p.getHandlerByFuncName(name)
	if err != nil {
		return nil, err
	}

	handlerCtx := context.HandlerContext{
		RegCtx:             p.ctx,
		RegisteredHandler:  handlerFunc,
		ExistingVarMap:     make(map[*types.Var]bool),
		ResolvedAssignExpr: make(map[string]string),
	}

	ri := returninspector.NewReturnInspector(handlerCtx)
	pi := payloadinspector.NewPayloadInspector(handlerCtx)
	inspectorList := []inspector.Inspector{
		ri, pi,
	}

	ast.Inspect(handlerFunc.FuncDecl, func(n ast.Node) bool {
		for _, inspector := range inspectorList {
			inspector.Inspect(n)
		}
		return true
	})
	handlerFunc.PayloadInfo = pi.Results
	handlerFunc.ReturnResponse = ri.Results
	return handlerFunc, nil
}
