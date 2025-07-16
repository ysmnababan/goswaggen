package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"

	"golang.org/x/tools/go/packages"
)

func TryTraverse() {
	fset = token.NewFileSet()
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
		Fset: fset,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		panic(err)
	}
	if len(pkgs) == 0 {
		log.Println("no package found")
		return
	}

	mainPkg := pkgs[0]
	var mainFunc *ast.FuncDecl
	var mainFile *ast.File
	// find 'main.go' file
	for i, file := range mainPkg.Syntax {
		fmt.Println(mainPkg.GoFiles[i], "=>>")
		mainFunc = SearchFuncNode(file, "main")
		if mainFunc != nil {
			mainFile = file
			break
		}
	}
	fmwork := FindFrameworkImportIdentName(mainFile, "echo")
	if len(fmwork) == 0 {
		return
	}
	fmt.Println("FRAMEWORK import name", fmwork)

	callExp := FindFmWorkInitExpression(mainFile, fmwork, "New")
	if callExp == nil {
		fmt.Println("Framework initializer is not found")
		return
	}
	fmt.Println(callExp.Lhs[0])
	// httpserverImport, ok := mainPkg.Imports[""]
	// if !ok {
	// log.Println("echo library import not found")
	// return
	// }

	// fmt.Println(httpserverImport.GoFiles)
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

// FindFrameworkInitAssignment searches for an assignment like:
// e := <frameworkName>.<functionName>()
// For example: e := echo.New()
func FindFmWorkInitExpression(file *ast.File, frameworkName string, functionName string) *ast.AssignStmt {
	var result *ast.AssignStmt
	ast.Inspect(file, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || len(assign.Rhs) != 1 {
			return true
		}

		callExpr, ok := assign.Rhs[0].(*ast.CallExpr)
		if !ok {
			return true
		}

		selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := selExpr.X.(*ast.Ident)
		if !ok {
			return true
		}

		if ident.Name == frameworkName && selExpr.Sel.Name == functionName {
			result = assign
			return false // stop walking
		}

		return true
	})
	return result
}
