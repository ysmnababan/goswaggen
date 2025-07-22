package main

import (
	"go/ast"

	"golang.org/x/tools/go/packages"
)

type RegistrationContext struct {
	Pkgs             []*packages.Package
	GroupPath        map[string]string
	// CurrentPackage   *packages.Package
	CurrentExpr      *ast.CallExpr
	CurrentFunc      *ast.FuncDecl
	funcDeclToPkgMap map[*ast.FuncDecl]*packages.Package // for faster retrival of a particular package
}

func (c *RegistrationContext) BuildFuncDeclToPkgMap() {
	declToPkg := make(map[*ast.FuncDecl]*packages.Package)
	for _, pkg := range c.Pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					declToPkg[fn] = pkg
				}
			}
		}
	}
	c.funcDeclToPkgMap = declToPkg
}
func (c *RegistrationContext) GetCurrentPackage() *packages.Package {
	return c.funcDeclToPkgMap[c.CurrentFunc]
}
