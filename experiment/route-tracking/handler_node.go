package main

import (
	"go/ast"
	"go/types"
)

type HandlerRegistration struct {
	Func      *types.Func   // The resolved handler function
	Call      *ast.CallExpr // The actual call expression
	IsDirect  bool          // True if registered directly on Echo instance
	GroupPath []string      // Group path segments, e.g., ["/api", "/v1"]
	FullPath  string        // Combined path, e.g., "/api/v1/resource"
	FromFunc  *types.Func   // The function where this registration happens
}
