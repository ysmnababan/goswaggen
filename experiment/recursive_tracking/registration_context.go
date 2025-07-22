package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type RegistrationContext struct {
	Pkgs                  []*packages.Package
	GroupPath             map[string]string
	funcDeclToPkgMap      map[*ast.FuncDecl]*packages.Package // cache for faster retrival of a particular package
	typeFuncToFuncDeclMap map[*types.Func]*ast.FuncDecl       // cache for faster retrival of a function declaration
	CurrentExpr           *ast.CallExpr
	CurrentFunc           *ast.FuncDecl

	// If a function use an `echo.Group` as an argument,
	// the variable itself must have actual path which is defined when the function is used.
	// So before peek into the `ast.FuncDecl`, save the actual path first.
	// Remember to reset after using so it doesn't clutter for the next function call.
	AliasForRouterTypeArgs string
}

func NewRegistrationContext(pkgs []*packages.Package, funDecl *ast.FuncDecl) *RegistrationContext {
	ctx := &RegistrationContext{
		funcDeclToPkgMap:      make(map[*ast.FuncDecl]*packages.Package),
		typeFuncToFuncDeclMap: make(map[*types.Func]*ast.FuncDecl),
		GroupPath:             make(map[string]string),
		Pkgs:                  pkgs,
		CurrentFunc:           funDecl,
	}
	ctx.buildFuncCache()
	return ctx
}

func (c *RegistrationContext) buildFuncCache() {
	declToPkg := make(map[*ast.FuncDecl]*packages.Package)
	typeFuncToFuncDeclMap := make(map[*types.Func]*ast.FuncDecl)
	for _, pkg := range c.Pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					declToPkg[fn] = pkg
					if obj, ok := pkg.TypesInfo.Defs[fn.Name]; ok {
						if fnObj, ok := obj.(*types.Func); ok {
							typeFuncToFuncDeclMap[fnObj] = fn
						}
					}
				}
			}
		}
	}
	c.funcDeclToPkgMap = declToPkg
	c.typeFuncToFuncDeclMap = typeFuncToFuncDeclMap
}

func (c *RegistrationContext) GetCurrentPackage() *packages.Package {
	return c.funcDeclToPkgMap[c.CurrentFunc]
}

func (c *RegistrationContext) GetFuncDecl(fnObj *types.Func) *ast.FuncDecl {
	return c.typeFuncToFuncDeclMap[fnObj]
}
