package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"strings"

	"golang.org/x/tools/go/packages"
)

var registrationHandler = []func(ctx *RegistrationContext) (*HandlerRegistration, bool){
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
		// fmt.Printf("%v IS REGISTERED IN %v >>>>> \n\n", val.Call.Fun, val.Func)
		val.Print()
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
func handleDirectRegistration(ctx *RegistrationContext) (*HandlerRegistration, bool) {
	mainPkg := ctx.MainPkg
	exp := ctx.Expr
	t, ok := exp.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, false
	}
	x, ok := t.X.(*ast.Ident)
	if !ok {
		return nil, false
	}
	obj, ok := mainPkg.TypesInfo.Uses[x]
	if !ok {
		return nil, false
	}
	if obj.Type().String() != ECHO_VARIABLE_TYPE {
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
		Func:     fn,
		Call:     exp,
		IsDirect: true,
	}
	return out, true
}

func FindFuncDeclaration(p *packages.Package, fnObj *types.Func) {
	for _, file := range p.Syntax {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			obj, ok := p.TypesInfo.Defs[fn.Name]
			if !ok {
				continue
			}
			if obj == fnObj {
				fmt.Println("MATCH: ", obj)
			}
		}
	}
}

// target pattern that can be recognized:
// RegisterEcho(e, "ignore-this")
func handleFunctionRegistration(ctx *RegistrationContext) (*HandlerRegistration, bool) {
	mainPkg := ctx.MainPkg
	exp := ctx.Expr
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
		Func:     fn,
		Call:     exp,
		IsDirect: false,
	}
	FindFuncDeclaration(mainPkg, fn)
	return out, true
}

func handleImportedFunctionRegistration(ctx *RegistrationContext) (*HandlerRegistration, bool) {
	return nil, false
}

// handleGroupRegistration
// search for registration with path like this:
// second_group := first_group.Group("/second")
// second_group.GET("/lol2", HandlerForSecondGroup)
func handleGroupRegistration(ctx *RegistrationContext) (*HandlerRegistration, bool) {
	mainPkg := ctx.MainPkg
	exp := ctx.Expr
	t, ok := exp.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, false
	}
	x, ok := t.X.(*ast.Ident)
	if !ok {
		return nil, false
	}
	obj, ok := mainPkg.TypesInfo.Uses[x]
	if !ok {
		return nil, false
	}
	if obj.Type().String() != ECHO_GROUP_VARIABLE_TYPE {
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
	path := ctx.GroupPath[obj.Name()]
	out := &HandlerRegistration{
		Func:     fn,
		Call:     exp,
		IsDirect: true,
		BasePath: path,
	}
	return out, true
}

// find all simple handler registration
// only contains the type of function (not the ast node)
// need to be inspected later on
// pattern that can be recognized:
// e.GET("/next-test", handlerTest)
// e.POST("/dummy", dummyhandler.JustDummyHandler)
func FindHandlerRegistrationNode(mainPkg *packages.Package, file *ast.File, identName string) []*HandlerRegistration {
	out := []*HandlerRegistration{}
	ctx := RegistrationContext{
		MainPkg:   mainPkg,
		GroupPath: make(map[string]string),
	}
	ast.Inspect(file, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.CallExpr: // this is for registration with `echo.echo`
			if len(t.Args) < 1 {
				return true
			}
			ctx.Expr = t
			for _, fn := range registrationHandler {
				handlerReg, ok := fn(&ctx)
				if ok {
					out = append(out, handlerReg)
					return false
				}
			}
			return false
		case *ast.AssignStmt: // this is for registration for `echo.Group`
			if len(t.Lhs) > 1 || len(t.Rhs) > 1 {
				return true
			}
			lhs, ok := t.Lhs[0].(*ast.Ident)
			if !ok {
				return true
			}
			callExpr, ok := t.Rhs[0].(*ast.CallExpr)
			if !ok {
				return true
			}

			if len(callExpr.Args) == 0 {
				return true
			}

			route, ok := callExpr.Args[0].(*ast.BasicLit)
			if !ok {
				return true
			}
			selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if selExpr.Sel.Name != "Group" {
				return true
			}
			ident, ok := selExpr.X.(*ast.Ident)
			if !ok {
				return true
			}
			if obj, ok := mainPkg.TypesInfo.Uses[ident]; ok {
				if obj.Type().String() == ECHO_VARIABLE_TYPE {
					ctx.GroupPath[lhs.Name] = strings.Trim(route.Value, `"`)
					return false
				}
				if obj.Type().String() == ECHO_GROUP_VARIABLE_TYPE {
					parentPath := ctx.GroupPath[ident.Name]
					ctx.GroupPath[lhs.Name] = parentPath + strings.Trim(route.Value, `"`)
					return false
				}
				return true
			}
			return true
		default:
			return true
		}
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
