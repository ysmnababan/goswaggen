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
	for id, obj := range mainPkg.TypesInfo.Uses {
		if obj != nil {
			fmt.Printf("%s: %q uses %v\n", cfg.Fset.Position(id.Pos()), id.Name, obj)
		}
	}
	fmwork := FindFrameworkImportIdentName(mainFile, "echo")
	if len(fmwork) == 0 {
		return
	}
	fmt.Println("FRAMEWORK import name", fmwork)

	initCallExp, identName := FindFmWorkInitExpression(mainFile, fmwork, "New")
	if initCallExp == nil {
		fmt.Println("Framework initializer is not found")
		return
	}
	fmt.Println("identName: ", identName)
	callExps := FindHandlerRegistrationNode(mainFile, identName)
	if len(callExps) == 0 {
		fmt.Println("can't find hander registration")
		return
	}
	if ident, ok := callExps[0].Args[1].(*ast.Ident); ok {
		obj := mainPkg.TypesInfo.Uses[ident]
		fmt.Println("hello", obj)
	}
	// fmt.Println(callExps[0])
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
// output => assign stmt and 'e'
func FindFmWorkInitExpression(file *ast.File, frameworkName string, functionName string) (*ast.AssignStmt, string) {
	var result *ast.AssignStmt
	var identName string
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
			identName = assign.Lhs[0].(*ast.Ident).Name
			return false // stop walking
		}

		return true
	})
	return result, identName
}

// find all simple handler registration
// i.e e.GET("/next-test", handlerTest)
func FindHandlerRegistrationNode(file *ast.File, identName string) []*ast.CallExpr {
	result := []*ast.CallExpr{}
	ast.Inspect(file, func(n ast.Node) bool {
		callExp, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if len(callExp.Args) < 1 {
			return true
		}
		selExp, ok := callExp.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if x, ok := selExp.X.(*ast.Ident); !ok || x.Name != identName {
			return true
		}

		if !IsHTTPMethod(selExp.Sel.Name) {
			return true
		}
		if _, ok := callExp.Args[0].(*ast.BasicLit); !ok {
			return true
		}
		var ident *ast.Ident
		if ident, ok = callExp.Args[1].(*ast.Ident); !ok {
			return true
		}
		fmt.Println(ident.Name)
		result = append(result, callExp)

		// obj := types.Info.Uses[ident]
		// _ = obj
		return false
	})
	return result
}

func IsHTTPMethod(method string) bool {
	return method == "GET" ||
		method == "POST" ||
		method == "PUT" ||
		method == "DELETE" ||
		method == "PATCH" ||
		method == "HEAD"
}
