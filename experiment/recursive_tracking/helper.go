package main

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Resolve the func object
// i.e: e.GET("/next-test", handler)
// the 'handler' param can be direct handler or '<somepackage>.handler
func resolveHandlerExpr(pkg *packages.Package, expr ast.Expr) (*types.Func, bool) {
	var fn *types.Func
	switch t := expr.(type) {
	case *ast.Ident: // i.e: e.GET("/next-test", handlerTest)
		obj, ok := pkg.TypesInfo.Uses[t]
		if !ok {
			return nil, false
		}
		fn, ok = obj.(*types.Func)
		if !ok {
			return nil, false
		}
	case *ast.SelectorExpr: // i.e: e.POST("/dummy", dummyhandler.JustDummyHandler)
		obj, ok := pkg.TypesInfo.Uses[t.Sel]
		if !ok {
			return nil, false
		}
		fn, ok = obj.(*types.Func)
		if !ok {
			return nil, false
		}
	default:
		return nil, false
	}
	return fn, true
}

func IsHTTPMethod(method string) bool {
	return method == "GET" ||
		method == "POST" ||
		method == "PUT" ||
		method == "DELETE" ||
		method == "PATCH" ||
		method == "HEAD"
}

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

func FindLibrary(imports []*ast.ImportSpec, targetName string) *ast.ImportSpec {
	path := IMPORT_PATH_VALUE[targetName]
	for _, val := range imports {
		if strings.Contains(val.Path.Value, path) {
			return val
		}
	}
	return nil
}

// finding a name used for importing a particular framework
// i.e  import ef "github.com/labstack/echo/v4"
// output=> 'ef'
func FindFrameworkImportIdentName(file *ast.File, fmworkName string) string {
	var fmworkIdent string
	fmworkImport := FindLibrary(file.Imports, fmworkName)
	if fmworkImport != nil {
		if fmworkImport.Name != nil {
			fmworkIdent = fmworkImport.Name.Name
		} else {
			fmworkIdent = "echo"
		}
	}
	return fmworkIdent
}
