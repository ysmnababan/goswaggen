package main

import (
	"go/ast"
	"go/types"
	"log"

	"golang.org/x/tools/go/packages"
)

type RegistrationContext struct {
	Pkgs                  []*packages.Package
	GroupPath             map[string]string
	funcDeclToPkgMap      map[*ast.FuncDecl]*packages.Package // cache for faster retrival of a particular package
	typeFuncToFuncDeclMap map[*types.Func]*ast.FuncDecl       // cache for faster retrival of a function declaration
	typeVarToGenDeclMap   map[*types.TypeName]*ast.GenDecl    // cache for faster retrival of a generic declaration
	CurrentExpr           *ast.CallExpr
	CurrentFunc           *ast.FuncDecl
	Level                 int

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
		typeVarToGenDeclMap:   make(map[*types.TypeName]*ast.GenDecl),
		GroupPath:             make(map[string]string),
		Pkgs:                  pkgs,
		CurrentFunc:           funDecl,
	}
	ctx.buildCache()
	return ctx
}

func (c *RegistrationContext) buildCache() {
	declToPkg := make(map[*ast.FuncDecl]*packages.Package)
	typeFuncToFuncDeclMap := make(map[*types.Func]*ast.FuncDecl)
	typeVarToGenDeclMap := make(map[*types.TypeName]*ast.GenDecl)
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
				if gn, ok := decl.(*ast.GenDecl); ok {
					if typeSpec, ok := gn.Specs[0].(*ast.TypeSpec); ok {
						if obj, ok := pkg.TypesInfo.Defs[typeSpec.Name]; ok {
							if varObj, ok := obj.(*types.TypeName); ok {
								typeVarToGenDeclMap[varObj] = gn
							}
						}
					}
				}
			}
		}
	}
	c.funcDeclToPkgMap = declToPkg
	c.typeFuncToFuncDeclMap = typeFuncToFuncDeclMap
	c.typeVarToGenDeclMap = typeVarToGenDeclMap
}

func (c *RegistrationContext) GetCurrentPackage() *packages.Package {
	return c.funcDeclToPkgMap[c.CurrentFunc]
}

func (c *RegistrationContext) GetFuncDecl(fnObj *types.Func) *ast.FuncDecl {
	out, ok := c.typeFuncToFuncDeclMap[fnObj]
	if !ok {
		log.Print(fnObj.FullName())
	}
	return out
}

func (c *RegistrationContext) ResetGroupPath() {
	for k := range c.GroupPath {
		delete(c.GroupPath, k)
	}
}

func (c *RegistrationContext) CreateGroupPathDuplicate() map[string]string {
	// Create a new map
	copyMap := make(map[string]string)

	for k, v := range c.GroupPath {
		copyMap[k] = v
	}
	return copyMap
}
