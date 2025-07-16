package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/packages"
)

func TryFindMainFile() {
	fset = token.NewFileSet()
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			// packages.NeedCompiledGoFiles |
			// packages.NeedImports |
			// packages.NeedDeps |
			// packages.NeedTypes |
			packages.NeedModule |
			packages.NeedSyntax,
		// packages.NeedTypesInfo,
		Dir:  "../example/learn-go/", // relative to where you run `go run`
		Fset: fset,
	}
	pkgs, err := packages.Load(cfg)
	if err != nil {
		panic(err)
	}

	var mainFile *ast.FuncDecl
	var imports []*ast.ImportSpec = nil

	for _, pkg := range pkgs {
		// each package contains many file, i.e main.go
		for i, file := range pkg.Syntax {
			fmt.Println(pkg.GoFiles[i], "=>>")
			mainFile = SearchFuncNode(file, "main")
			if mainFile != nil {
				imports = file.Imports
				break
			}
		}
	}

	var echoIdent string
	echoImport := FindLibrary(imports, "github.com/labstack/echo/v4")
	if echoImport != nil {
		if echoImport.Name != nil {
			echoIdent = echoImport.Name.Name
		} else {
			echoIdent = "echo"
		}
	}
	fmt.Println("echo import name", echoIdent)
}

func FindLibrary(imports []*ast.ImportSpec, targetName string) *ast.ImportSpec {
	for _, val := range imports {
		if strings.Contains(val.Path.Value, targetName) {
			return val
		}
	}
	return nil
}
func SearchFuncNode(f *ast.File, targetName string) *ast.FuncDecl {
	var fnode *ast.FuncDecl = nil
	ast.Inspect(f, func(n ast.Node) bool {
		fun, ok := n.(*ast.FuncDecl)
		if ok && fun.Name.Name == targetName {
			fnode = fun
			return false
		}
		return true
	})
	return fnode
}
