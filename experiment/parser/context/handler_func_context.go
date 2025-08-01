package context

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

type HandlerContext interface {
	GetPackage() *packages.Package
	GetTypesInfo() *types.Info
	GetTypeNameToGenDeclCache() map[*types.TypeName]*ast.GenDecl
	GetTypePackageCache() map[*types.Package]*packages.Package
	GetTypeGlobalVarToValueMap() map[string]string
}
