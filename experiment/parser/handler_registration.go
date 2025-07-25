package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
)

type HandlerRegistration struct {
	Func      *types.Func   // The resolved handler function
	Call      *ast.CallExpr // The actual call expression
	IsDirect  bool          // True if registered directly on Echo instance
	GroupPath string        // Group path segments, e.g., ["/api", "/v1"]
	BasePath  string        // Combined path, e.g., "/api/v1/resource"
	FromFunc  *types.Func   // The function where this registration happens
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
