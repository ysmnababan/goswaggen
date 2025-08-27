package helper

import (
	"go/ast"

	"golang.org/x/tools/go/packages"
)

var MAIN_PACKAGE_NAME = "main"

func SearchDeclFun(pkgs []*packages.Package, targetName string, packageName *string) (*ast.FuncDecl, *ast.File) {
	for _, pkg := range pkgs {
		if packageName != nil && *packageName != "" && pkg.Name != *packageName {
			// search for particular package name (is requested)
			continue
		}
		for _, file := range pkg.Syntax {
			for _, funDecl := range file.Decls {
				fn, ok := funDecl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				if fn.Name.String() == targetName {
					return fn, file
				}
			}
		}
	}
	return nil, nil
}
