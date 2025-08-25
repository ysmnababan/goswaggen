package context

import (
	"go/ast"
	"go/types"

	"github.com/ysmnababan/goswaggen/internal/model"
	"golang.org/x/tools/go/packages"
)

type HandlerContext struct {
	RegCtx             *RegistrationContext
	BindArgExpr        *ast.Expr
	ExistingVarMap     map[*types.Var]bool
	RegisteredHandler  *model.HandlerRegistration
	ResolvedAssignExpr map[string]string
}

func (c *HandlerContext) GetPackage() *packages.Package {
	return c.RegisteredHandler.Pkg
}

func (c *HandlerContext) GetTypesInfo() *types.Info {
	return c.RegisteredHandler.Pkg.TypesInfo
}

func (c *HandlerContext) GetTypeNameToGenDeclCache() map[*types.TypeName]*ast.GenDecl {
	return c.RegCtx.typeVarToGenDeclMap
}

func (c *HandlerContext) GetTypePackageCache() map[*types.Package]*packages.Package {
	return c.RegCtx.packageMap
}

func (c *HandlerContext) GetTypeGlobalVarToValueMap() map[string]string {
	return c.RegCtx.typeGlobalVarToValueMap
}
