package main

import (
	"go/ast"
	"go/token"
	"go/types"
	"log"

	"golang.org/x/tools/go/packages"
)

type RegistrationContext struct {
	Pkgs                    []*packages.Package
	GroupPath               map[string]string
	funcDeclToPkgMap        map[*ast.FuncDecl]*packages.Package // cache for faster retrival of a particular package
	typeFuncToFuncDeclMap   map[*types.Func]*ast.FuncDecl       // cache for faster retrival of a function declaration
	typeVarToGenDeclMap     map[*types.TypeName]*ast.GenDecl    // cache for faster retrival of a typeSpec generic declaration
	typeGlobalVarToValueMap map[string]string                   // cache for string global variable with it respected value
	CurrentExpr             *ast.CallExpr
	CurrentFunc             *ast.FuncDecl
	Level                   int

	// If a function use an `echo.Group` as an argument,
	// the variable itself must have actual path which is defined when the function is used.
	// So before peek into the `ast.FuncDecl`, save the actual path first.
	// Remember to reset after using so it doesn't clutter for the next function call.
	AliasForRouterTypeArgs string
}

func NewRegistrationContext(pkgs []*packages.Package, funDecl *ast.FuncDecl) *RegistrationContext {
	ctx := &RegistrationContext{
		funcDeclToPkgMap:        make(map[*ast.FuncDecl]*packages.Package),
		typeFuncToFuncDeclMap:   make(map[*types.Func]*ast.FuncDecl),
		typeVarToGenDeclMap:     make(map[*types.TypeName]*ast.GenDecl),
		typeGlobalVarToValueMap: make(map[string]string),

		GroupPath:   make(map[string]string),
		Pkgs:        pkgs,
		CurrentFunc: funDecl,
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
					switch spec := gn.Specs[0].(type) {
					case *ast.TypeSpec: // store the `type struct`
						if obj, ok := pkg.TypesInfo.Defs[spec.Name]; ok {
							if varObj, ok := obj.(*types.TypeName); ok {
								typeVarToGenDeclMap[varObj] = gn
							}
						}
					case *ast.ValueSpec: // store the `global var or constant`
						if obj, ok := pkg.TypesInfo.Defs[spec.Names[0]]; ok {
							if len(spec.Values) != 1 {
								continue
							}
							if val, ok := spec.Values[0].(*ast.BasicLit); ok && val.Kind == token.STRING {
								c.typeGlobalVarToValueMap[obj.String()] = val.Value
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
