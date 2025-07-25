package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

type HandlerRegistration struct {
	Func      *types.Func   // The resolved handler function as obj
	Call      *ast.CallExpr // The actual call expression
	IsDirect  bool          // True if registered directly on Framework instance
	GroupPath string        // Group path segments, e.g., ["/api", "/v1"]
	BasePath  string        // Combined path, e.g., "/api/v1/resource"
	FromFunc  *types.Func   // The function where this registration happens
	Pkg       *packages.Package
	FuncDecl  *ast.FuncDecl // The implementation of the handler function

	//
	Request []*RequestData // The requested data can be from body payload, query param or param
}

func (n *HandlerRegistration) Print() {
	fmt.Println(">>>>>>")
	if n.IsDirect {
		method := n.Call.Fun.(*ast.SelectorExpr)
		pathArg := n.Call.Args[0].(*ast.BasicLit)
		fullpath := `"` + n.BasePath + strings.Trim(pathArg.Value, `"`) + `"`
		fmt.Printf("%s.%s(%s,%s)\n", method.X, method.Sel.Name, fullpath, n.Func.Name())
	} else {
		method := n.Call.Fun.(*ast.Ident)
		fmt.Printf("%s(%v)\n", method.String(), n.Call.Args)
	}
	fmt.Println()
}

func (n *HandlerRegistration) GetFuncName() string {
	return n.Func.Name()
}

func (n *HandlerRegistration) GetMethod() string {
	method := n.Call.Fun.(*ast.SelectorExpr)
	return method.Sel.Name
}

func (n *HandlerRegistration) GetFullPath() string {
	pathArg := n.Call.Args[0].(*ast.BasicLit)
	return n.BasePath + strings.Trim(pathArg.Value, `"`)
}

type RequestData struct {
	Call       *ast.CallExpr // The actual call expression
	Param      *types.Var    // The param type
	ParamDecl  *ast.GenDecl  // Declaration of the param
	BindMethod string        // Body, QueryParam, Param
	BasicLit   string        // for queryparam and param args, e.g. <context>.Param("this")
}
