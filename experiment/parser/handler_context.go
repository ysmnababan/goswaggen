package main

import (
	"go/ast"
	"go/types"
)

type HandlerContext struct {
	RegCtx            *RegistrationContext
	BindArgExpr       *ast.Expr
	ExistingVarMap    map[*types.Var]bool
	RegisteredHandler *HandlerRegistration
}
