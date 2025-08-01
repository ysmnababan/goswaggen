package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type HandlerContext struct {
	RegCtx             *RegistrationContext
	BindArgExpr        *ast.Expr
	ExistingVarMap     map[*types.Var]bool
	RegisteredHandler  *HandlerRegistration
	ResolvedAssignExpr map[string]string
}

func (c *HandlerContext) GetPackage() *packages.Package {
	return c.RegisteredHandler.Pkg
}
