package main

import (
	"go/ast"

	"golang.org/x/tools/go/packages"
)

type RegistrationContext struct {
	MainPkg   *packages.Package
	Expr      *ast.CallExpr
	IdentName string
	GroupPath map[string]string

	// Add more fields as needed
}
