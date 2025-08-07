package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"parser/inspector"
	"parser/inspector/payloadinspector"
	"parser/inspector/returninspector"

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
	fmt.Println(handlerFunc.GetFuncName())
	fmt.Println(handlerFunc.GetFullPath())
	fmt.Println(handlerFunc.GetMethod())
	handlerCtx := &HandlerContext{
		RegCtx:             ctx,
		RegisteredHandler:  handlerFunc,
		ExistingVarMap:     make(map[*types.Var]bool),
		ResolvedAssignExpr: make(map[string]string),
	}
	ExtractFuncHandlerInfoRefactored(handlerCtx)
}

func ExtractFuncHandlerInfoRefactored(ctx *HandlerContext) {
	ri := returninspector.NewReturnInspector(ctx)
	pi := payloadinspector.NewPayloadInspector(ctx)
	inspectorList := []inspector.Inspector{
		ri, pi,
	}

	ast.Inspect(ctx.RegisteredHandler.FuncDecl, func(n ast.Node) bool {
		for _, inspector := range inspectorList {
			inspector.Inspect(n)
		}
		return true
	})

	ri.PrintResult()
	pi.PrintResult()
	ctx.RegisteredHandler.PayloadInfo = pi.Results
	ctx.RegisteredHandler.ReturnResponse = ri.Results
}
