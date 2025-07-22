package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strings"

	"golang.org/x/tools/go/packages"
)

var FSET *token.FileSet

func SearchDeclFun(pkgs []*packages.Package, targetName string, packageName *string) (*ast.FuncDecl, *ast.File) {
	for _, pkg := range pkgs {
		if packageName != nil && *packageName != "" && pkg.Name != *packageName {
			// search for particular package name (is requested)
			continue
		}
		for _, file := range pkg.Syntax {
			for _, funDecl := range file.Decls {
				fn, ok := funDecl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				if fn.Name.String() == targetName {
					return fn, file
				}
			}
		}
	}
	return nil, nil
}

func FindLibrary(imports []*ast.ImportSpec, targetName string) *ast.ImportSpec {
	path := IMPORT_PATH_VALUE[targetName]
	for _, val := range imports {
		if strings.Contains(val.Path.Value, path) {
			return val
		}
	}
	return nil
}

// finding a name used for importing a particular framework
// i.e  import ef "github.com/labstack/echo/v4"
// output=> 'ef'
func FindFrameworkImportIdentName(file *ast.File, fmworkName string) string {
	var fmworkIdent string
	fmworkImport := FindLibrary(file.Imports, fmworkName)
	if fmworkImport != nil {
		if fmworkImport.Name != nil {
			fmworkIdent = fmworkImport.Name.Name
		} else {
			fmworkIdent = "echo"
		}
	}
	return fmworkIdent
}

func TryRecursive() {
	FSET = token.NewFileSet()
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
		Dir:  "../example/learn-go/", // relative to where you run `go run`
		Fset: FSET,
	}
	pkgs, err := packages.Load(cfg, ".") // load add the package
	if err != nil {
		panic(err)
	}
	if len(pkgs) == 0 {
		log.Println("no package found")
		return
	}

	mainPackageName := "main"
	mainFuncDecl, mainFile := SearchDeclFun(pkgs, "main", &mainPackageName)
	if mainFuncDecl == nil {
		log.Println("no main file found")
		return
	}

	fmwork := FindFrameworkImportIdentName(mainFile, "echo")
	if len(fmwork) == 0 {
		return
	}
	fmt.Println("FRAMEWORK import name", fmwork)
}
