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
	handlerCtx := &HandlerContext{
		RegCtx:            ctx,
		RegisteredHandler: handlerFunc,
		ExistingVarMap:    make(map[*types.Var]bool),
	}
	out := SearchBindRequest(handlerCtx)
	if len(out) == 0 {
		log.Println("no requested data ")
		return
	}

	for _, req := range out {
		param := ""
		if req.Param != nil {
			param = req.Param.String()
		}
		fmt.Printf("method: %s, param: %v, basicLit: %s\n", req.BindMethod, param, req.BasicLit)
		if req.ParamDecl != nil {
			fmt.Println(req.ParamDecl.Specs[0])
		}
	}
}

func SearchBindRequest(ctx *HandlerContext) []*RequestData {
	var result []*RequestData
	ast.Inspect(ctx.RegisteredHandler.FuncDecl, func(n ast.Node) bool {
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
		ident, ok := ctx.RegisteredHandler.Pkg.TypesInfo.Uses[x]
		if !ok {
			return true
		}

		if ident.Type().String() != ECHO_CONTEXT_TYPE {
			return true
		}
		bindMethod := selExpr.Sel.Name
		reqData := &RequestData{}

		// extract `&req` => `req`
		argExp := callExpr.Args[0]
		if unaryExp, ok := callExpr.Args[0].(*ast.UnaryExpr); ok && unaryExp.Op == token.AND {
			argExp = unaryExp.X
		}
		ctx.BindArgExpr = &argExp
		switch bindMethod {
		case "Bind":
			if reqData, ok = resolveBind(ctx); !ok {
				return true
			}
		case "QueryParam":
			reqData, ok = resolveQueryParam(ctx)
			if !ok {
				return true
			}
		case "Param":
			reqData, ok = resolveParam(ctx)
			if !ok {
				return true
			}
		default:
			return true
		}
		reqData.Call = callExpr
		reqData.BindMethod = bindMethod
		result = append(result, reqData)
		return false
	})

	return result
}

func resolveBind(ctx *HandlerContext) (*RequestData, bool) {
	argExp := ctx.BindArgExpr
	h := ctx.RegisteredHandler
	objMap := ctx.ExistingVarMap
	regCtx := ctx.RegCtx
	var paramIdent *ast.Ident
	switch exp := (*argExp).(type) {
	case *ast.Ident:
		paramIdent = exp
	case *ast.SelectorExpr:
		// for case like `<c>.Bind(<some-struct>.<Selector>)`,
		// extract the <Selector> first
		paramIdent = exp.Sel
	default:
		return nil, false
	}

	obj, ok := h.Pkg.TypesInfo.Uses[paramIdent]
	if !ok {
		return nil, false
	}
	v, ok := obj.(*types.Var)
	if !ok || objMap[v] {
		return nil, false

	}
	reqData := new(RequestData)
	reqData.Param = v
	// Get underlying struct type
	typ := v.Type()
	if ptr, ok := typ.(*types.Pointer); ok {
		typ = ptr.Elem()
	}
	named, ok := typ.(*types.Named)
	if !ok {
		return nil, false
	}
	typeName := named.Obj() // *types.TypeName
	if decl, ok := regCtx.typeVarToGenDeclMap[typeName]; ok {
		reqData.ParamDecl = decl
	}
	objMap[v] = true
	return reqData, true
}

func resolveQueryParam(ctx *HandlerContext) (*RequestData, bool) {
	argExp := ctx.BindArgExpr
	h := ctx.RegisteredHandler
	objMap := ctx.ExistingVarMap
	regCtx := ctx.RegCtx
	switch arg := (*argExp).(type) {
	case *ast.Ident:
		obj, ok := h.Pkg.TypesInfo.Uses[arg]
		if !ok {
			return nil, false
		}
		reqData := &RequestData{}
		switch v := obj.(type) {
		case *types.Var:
			if objMap[v] {
				return nil, false
			}
			reqData.Param = v
			objMap[v] = true
			if basicLit, ok := regCtx.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
				return reqData, true
			}
		case *types.Const:
			if basicLit, ok := regCtx.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
				return reqData, true
			}
		default:
			return nil, false
		}
	case *ast.BasicLit:
		return &RequestData{BasicLit: arg.Value}, true
	default:
		fmt.Println("def: ", arg)
		return nil, false
	}
	return nil, false
}

func resolveParam(ctx *HandlerContext) (*RequestData, bool) {
	argExp := ctx.BindArgExpr
	h := ctx.RegisteredHandler
	objMap := ctx.ExistingVarMap
	regCtx := ctx.RegCtx
	switch arg := (*argExp).(type) {
	case *ast.Ident:
		obj, ok := h.Pkg.TypesInfo.Uses[arg]
		if !ok {
			return nil, false
		}
		reqData := &RequestData{}
		switch v := obj.(type) {
		case *types.Var:
			if objMap[v] {
				return nil, false
			}
			reqData.Param = v
			objMap[v] = true
			if basicLit, ok := regCtx.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
				return reqData, true
			}
		case *types.Const:
			if basicLit, ok := regCtx.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
				return reqData, true
			}
		default:
			return nil, false
		}
	case *ast.BasicLit:
		return &RequestData{BasicLit: arg.Value}, true
	default:
		fmt.Println("def: ", arg)
		return nil, false
	}
	return nil, false
}
