package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"

	"golang.org/x/tools/go/packages"
)

func TryParseHandler() {
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
		Dir:  "../example/learn-go/", // relative to where you run `go run`
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
	fmt.Println(len(handlerRegs))
	targetHandler := "CreateUser"
	var handlerFunc *HandlerRegistration
	for _, val := range handlerRegs {
		if val.Func.Name() == targetHandler {
			handlerFunc = val
		}
	}

	if handlerFunc == nil {
		log.Printf("handler with name '%s' was not found\n", targetHandler)
		return
	}
	fmt.Println(handlerFunc.Func.Name())
	fmt.Println(handlerFunc.GetFullPath())
	fmt.Println(handlerFunc.GetMethod())
	out := SearchBindRequest(ctx, handlerFunc)
	if len(out) == 0 {
		log.Println("no requested data ")
		return
	}

	for _, req := range out {
		fmt.Println(req.BindMethod, req.Param, req.BasicLit)
		if req.ParamDecl != nil {
			fmt.Println(req.ParamDecl.Specs[0])
		}
	}
}

func SearchBindRequest(ctx *RegistrationContext, h *HandlerRegistration) []*RequestData {
	var result []*RequestData
	objMap := make(map[*types.Var]bool)
	ast.Inspect(h.FuncDecl, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if len(callExpr.Args) != 1 {
			return true
		}

		selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		x, ok := selExpr.X.(*ast.Ident)
		if !ok {
			return true
		}
		ident, ok := h.Pkg.TypesInfo.Uses[x]
		if !ok {
			return true
		}

		if ident.Type().String() != ECHO_CONTEXT_TYPE {
			return true
		}
		bindMethod := selExpr.Sel.Name
		reqData := &RequestData{
			Call:       callExpr,
			BindMethod: bindMethod,
		}

		switch bindMethod {
		case "Bind":
			arg, ok := callExpr.Args[0].(*ast.Ident)
			if !ok {
				return true
			}
			obj, ok := h.Pkg.TypesInfo.Uses[arg]
			if !ok {
				return true
			}
			v, ok := obj.(*types.Var)
			if !ok || objMap[v] {
				return true
			}
			reqData.Param = v
			// Get underlying struct type
			typ := v.Type()
			if ptr, ok := typ.(*types.Pointer); ok {
				typ = ptr.Elem()
			}
			named, ok := typ.(*types.Named)
			if !ok {
				return true
			}
			typeName := named.Obj() // *types.TypeName
			if decl, ok := ctx.typeVarToGenDeclMap[typeName]; ok {
				reqData.ParamDecl = decl
			}
			objMap[v] = true
		case "QueryParam":
			switch arg := callExpr.Args[0].(type) {
			case *ast.Ident:
				obj, ok := h.Pkg.TypesInfo.Uses[arg]
				if !ok {
					return true
				}
				v, ok := obj.(*types.Var)
				if !ok || objMap[v] {
					return true
				}
				reqData.Param = v
				objMap[v] = true
			case *ast.BasicLit:
				reqData.BasicLit = arg.Value
			}
		case "Param":
			switch arg := callExpr.Args[0].(type) {
			case *ast.Ident:
				obj, ok := h.Pkg.TypesInfo.Uses[arg]
				if !ok {
					return true
				}
				v, ok := obj.(*types.Var)
				if !ok || objMap[v] {
					return true
				}
				reqData.Param = v
				objMap[v] = true
			case *ast.BasicLit:
				reqData.BasicLit = arg.Value
			}
		default:
			return true
		}
		result = append(result, reqData)
		return false
	})

	return result
}
