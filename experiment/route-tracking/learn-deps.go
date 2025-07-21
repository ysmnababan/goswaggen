package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"

	"golang.org/x/tools/go/packages"
)

var registrationHandler = []func(mainPkg *packages.Package, exp *ast.CallExpr, identName string) (*HandlerRegistration, bool){
	handleDirectRegistration,
	handleFunctionRegistration,
	handleImportedFunctionRegistration,
	handleGroupRegistration,
}

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
	// for id, obj := range mainPkg.TypesInfo.Uses {
	// 	if obj != nil {
	// 		fmt.Printf("%s: %q uses %v\n", cfg.Fset.Position(id.Pos()), id.Name, obj)
	// 	}
	// }
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
	handlerRegs := FindHandlerRegistrationNode(mainPkg, mainFile, identName)
	if len(handlerRegs) == 0 {
		fmt.Println("can't find handler registration")
		return
	}

	for _, val := range handlerRegs {
		fmt.Printf("%v IS REGISTERED IN %v >>>>> \n\n", val.Call.Fun, val.Func)
	}
	// check types of `ast.Ident`
	// if ident, ok := callExps[0].Args[1].(*ast.Ident); ok {
	// 	obj := mainPkg.TypesInfo.Uses[ident]
	// 	fmt.Printf("name:%v, type:%v, obj:%v, parent:%v\n", obj.Name(), obj.Type().String(), obj.Pkg(), obj.Parent())
	// }
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

// target pattern that can be recognized:
// e.GET("/next-test", handlerTest)
// e.POST("/dummy", dummyhandler.JustDummyHandler)
func handleDirectRegistration(mainPkg *packages.Package, exp *ast.CallExpr, identName string) (*HandlerRegistration, bool) {
	t, ok := exp.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, false
	}
	if x, ok := t.X.(*ast.Ident); !ok || x.Name != identName {
		return nil, false
	}
	if !IsHTTPMethod(t.Sel.Name) {
		return nil, false
	}
	if _, ok := exp.Args[0].(*ast.BasicLit); !ok {
		return nil, false
	}
	fn, ok := resolveHandlerExpr(mainPkg, exp.Args[1])
	if !ok {
		return nil, false
	}
	out := &HandlerRegistration{
		Func:      fn,
		Call:      exp,
		IsDirect:  true,
		GroupPath: []string{},
	}
	return out, true
}

// target pattern that can be recognized:
// RegisterEcho(e, "ignore-this")
func handleFunctionRegistration(mainPkg *packages.Package, exp *ast.CallExpr, identName string) (*HandlerRegistration, bool) {
	t, ok := exp.Fun.(*ast.Ident)
	if !ok {
		return nil, false
	}
	obj, ok := mainPkg.TypesInfo.Uses[t]
	if !ok {
		return nil, false

	}
	fn, ok := obj.(*types.Func)
	if !ok {
		return nil, false
	}
	out := &HandlerRegistration{
		Func:      fn,
		Call:      exp,
		IsDirect:  false,
		GroupPath: []string{},
	}
	return out, true
}

func handleImportedFunctionRegistration(mainPkg *packages.Package, exp *ast.CallExpr, identName string) (*HandlerRegistration, bool) {
	return nil, false
}

func handleGroupRegistration(mainPkg *packages.Package, exp *ast.CallExpr, identName string) (*HandlerRegistration, bool) {
	return nil, false
}

// find all simple handler registration
// only contains the type of function (not the ast node)
// need to be inspected later on
// pattern that can be recognized:
// e.GET("/next-test", handlerTest)
// e.POST("/dummy", dummyhandler.JustDummyHandler)
func FindHandlerRegistrationNode(mainPkg *packages.Package, file *ast.File, identName string) []*HandlerRegistration {
	out := []*HandlerRegistration{}
	ast.Inspect(file, func(n ast.Node) bool {
		callExp, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if len(callExp.Args) < 1 {
			return true
		}
		for _, fn := range registrationHandler {
			handlerReg, ok := fn(mainPkg, callExp, identName)
			if ok {
				out = append(out, handlerReg)
				return false
			}
		}
		return false
	})

	return out
}

// Resolve the func object
// i.e: e.GET("/next-test", handler)
// the 'handler' param can be direct handler or '<somepackage>.handler
func resolveHandlerExpr(pkg *packages.Package, expr ast.Expr) (*types.Func, bool) {
	var fn *types.Func
	switch t := expr.(type) {
	case *ast.Ident: // i.e: e.GET("/next-test", handlerTest)
		obj, ok := pkg.TypesInfo.Uses[t]
		if !ok {
			return nil, false
		}
		fn, ok = obj.(*types.Func)
		if !ok {
			return nil, false
		}
	case *ast.SelectorExpr: // i.e: e.POST("/dummy", dummyhandler.JustDummyHandler)
		obj, ok := pkg.TypesInfo.Uses[t.Sel]
		if !ok {
			return nil, false
		}
		fn, ok = obj.(*types.Func)
		if !ok {
			return nil, false
		}
	default:
		return nil, false
	}
	return fn, true
}

func IsHTTPMethod(method string) bool {
	return method == "GET" ||
		method == "POST" ||
		method == "PUT" ||
		method == "DELETE" ||
		method == "PATCH" ||
		method == "HEAD"
}
