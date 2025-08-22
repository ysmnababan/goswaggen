package parser

import (
	"fmt"
	"go/token"
	"log"

	"github.com/ysmnababan/goswaggen/internal/parser/context"
	"github.com/ysmnababan/goswaggen/internal/parser/tracking"
	"golang.org/x/tools/go/packages"
)

var FSET *token.FileSet
var MAIN_PACKAGE_NAME = "main"

func ParseHandler(dir string) {
	FSET = token.NewFileSet()
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			// packages.NeedCompiledGoFiles |
			packages.NeedImports |
			// packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
		Dir:  dir, // relative to where you run `go run`
		Fset: FSET,
	}

	pkgs, err := packages.Load(cfg, "./...") // load add the package
	if err != nil {
		panic(err)
	}
	if len(pkgs) == 0 {
		log.Println("no package found")
		return
	}
	mainFuncDecl, _ := searchDeclFun(pkgs, "main", &MAIN_PACKAGE_NAME)
	if mainFuncDecl == nil {
		log.Println("no main file found")
		return
	}

	ctx := context.NewRegistrationContext(pkgs, mainFuncDecl)
	handlerRegs := tracking.FindHandlerRegistration(ctx)
	if len(handlerRegs) == 0 {
		fmt.Println("can't find handler registration")
		return
	}
}
