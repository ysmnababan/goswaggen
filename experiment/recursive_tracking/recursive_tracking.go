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

var FSET *token.FileSet

var registrationHandler = []func(ctx *RegistrationContext) (*HandlerRegistration, bool){
	handleDirectRegistration,
	handleGroupRegistration,
	handleFunctionRegistration,
	// handleImportedFunctionRegistration,
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
	ctx := NewRegistrationContext(pkgs, mainFuncDecl)
	handlerRegs := FindHandlerRegistration(ctx)
	if len(handlerRegs) == 0 {
		fmt.Println("can't find handler registration")
		return
	}

	for _, val := range handlerRegs {
		val.Print()
	}
}

// target pattern that can be recognized:
// e.GET("/next-test", handlerTest)
// e.POST("/dummy", dummyhandler.JustDummyHandler)
func handleDirectRegistration(ctx *RegistrationContext) (*HandlerRegistration, bool) {
	pkg := ctx.GetCurrentPackage()
	exp := ctx.CurrentExpr
	t, ok := exp.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, false
	}
	x, ok := t.X.(*ast.Ident)
	if !ok {
		return nil, false
	}
	obj, ok := pkg.TypesInfo.Uses[x]
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
	fn, ok := resolveHandlerExpr(pkg, exp.Args[1])
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

// handleGroupRegistration
// search for registration with path like this:
// second_group := first_group.Group("/second")
// second_group.GET("/lol2", HandlerForSecondGroup)
func handleGroupRegistration(ctx *RegistrationContext) (*HandlerRegistration, bool) {
	pkg := ctx.GetCurrentPackage()
	exp := ctx.CurrentExpr
	t, ok := exp.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, false
	}
	x, ok := t.X.(*ast.Ident)
	if !ok {
		return nil, false
	}
	obj, ok := pkg.TypesInfo.Uses[x]
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
	fn, ok := resolveHandlerExpr(pkg, exp.Args[1])
	if !ok {
		return nil, false
	}
	path, ok := ctx.GroupPath[obj.Name()]
	if !ok {
		path = ctx.AliasForRouterTypeArgs
	}
	out := &HandlerRegistration{
		Func:     fn,
		Call:     exp,
		IsDirect: true,
		BasePath: path,
	}
	return out, true
}

// getRouterTypePrefix
//
// Checking if a function contains a router type prefix.
//
// case 1: foo(<echo>)
//
// case 2: foo(<echo.Group>)
//
// case 3: foo(<echo>.Group("/path"))
//
// case 4: foo(<echo.Group>.Group("/path"))
func getRouterTypePrefix(ctx *RegistrationContext) (bool, string) {
	pkg := ctx.GetCurrentPackage()
	for _, arg := range ctx.CurrentExpr.Args {
		switch n := arg.(type) {
		case *ast.Ident:
			obj, ok := pkg.TypesInfo.Uses[n]
			if !ok {
				continue
			}
			if obj.Type().String() == ECHO_VARIABLE_TYPE {
				return true, "" // case 1
			}
			if obj.Type().String() == ECHO_GROUP_VARIABLE_TYPE {
				return true, ctx.GroupPath[obj.Name()] // case 2
			}
		case *ast.CallExpr:
			if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					continue
				}
				obj, ok := pkg.TypesInfo.Uses[ident]
				if !ok {
					continue
				}

				if obj.Type().String() == ECHO_VARIABLE_TYPE {
					route, ok := n.Args[0].(*ast.BasicLit)
					if !ok {
						continue
					}
					return true, strings.Trim(route.Value, `"`) // case 3
				}
				if obj.Type().String() == ECHO_GROUP_VARIABLE_TYPE {
					prefix := ctx.GroupPath[obj.Name()]
					route, ok := n.Args[0].(*ast.BasicLit)
					if !ok {
						continue
					}
					prefix += strings.Trim(route.Value, `"`)
					return true, prefix // case 4
				}
			}
		default:
			continue
		}
	}
	return false, ""
}

// Target pattern that can be recognized:
//
// RegisterEcho(e, "ignore-this")
func handleFunctionRegistration(ctx *RegistrationContext) (*HandlerRegistration, bool) {
	pkg := ctx.GetCurrentPackage()
	exp := ctx.CurrentExpr

	// make sure the filter out the function that not using registration param like `echo.echo` or `echo.Group`
	hasRouterPrefix, prefix := getRouterTypePrefix(ctx)
	if !hasRouterPrefix {
		return nil, false
	}
	t, ok := exp.Fun.(*ast.Ident)
	if !ok {
		return nil, false
	}
	obj, ok := pkg.TypesInfo.Uses[t]
	if !ok {
		return nil, false

	}
	fn, ok := obj.(*types.Func)
	if !ok {
		return nil, false
	}
	ctx.AliasForRouterTypeArgs = prefix
	out := &HandlerRegistration{
		Func:     fn,
		Call:     exp,
		IsDirect: false,
	}
	return out, true
}

// find all simple handler registration
// only contains the type of function (not the ast node)
// need to be inspected later on
// pattern that can be recognized:
// e.GET("/next-test", handlerTest)
// e.POST("/dummy", dummyhandler.JustDummyHandler)
func FindHandlerRegistration(ctx *RegistrationContext) []*HandlerRegistration {
	var result []*HandlerRegistration

	ast.Inspect(ctx.CurrentFunc, func(n ast.Node) bool {
		switch t := n.(type) {
		case *ast.CallExpr: // this is for registration with `echo.echo`
			if len(t.Args) < 1 {
				return true
			}
			ctx.CurrentExpr = t
			for _, fn := range registrationHandler {
				handlerReg, ok := fn(ctx)
				if ok {
					if handlerReg.IsDirect {
						result = append(result, handlerReg)
						return false
					}

					// when the expression is not direct handler registration
					// have to traverse inside the FuncDecl again.
					ctx.CurrentExpr = nil
					ctx.CurrentFunc = ctx.GetFuncDecl(handlerReg.Func)
					regs := FindHandlerRegistration(ctx)
					ctx.AliasForRouterTypeArgs = "" // reset alias for each `ast.FuncDecl` inspect
					result = append(result, regs...)
					return false
				}
			}
			return false
		case *ast.AssignStmt: // this is for finding the Grouping
			currentPkg := ctx.GetCurrentPackage()
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
			if obj, ok := currentPkg.TypesInfo.Uses[ident]; ok {
				if obj.Type().String() == ECHO_VARIABLE_TYPE {
					ctx.GroupPath[lhs.Name] = strings.Trim(route.Value, `"`)
					return false
				}
				if obj.Type().String() == ECHO_GROUP_VARIABLE_TYPE {
					parentPath := ctx.AliasForRouterTypeArgs
					if parentPath == "" {
						parentPath = ctx.GroupPath[ident.Name]
					}
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

	return result
}

// TODO
// RegisterEchoSelectorAsParamGroup(first_group.Group("/"))
//
