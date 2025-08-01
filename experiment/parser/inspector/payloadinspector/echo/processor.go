package echo

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"parser/context"
	"parser/framework"
	"parser/model"
	"parser/tagparserutil"

	"golang.org/x/tools/go/packages"
)

type EchoPayloadProcessor struct {
	resolvedAssignExpr map[string]string
	// pkg                    *packages.Package
	typesInfo              *types.Info
	visitedVar             map[*types.Var]bool
	typeNameToGenDeclCache map[*types.TypeName]*ast.GenDecl

	// packageMap maps types.Package to its corresponding packages.Package,
	// allowing access to AST and type information across packages.
	typePackageCache map[*types.Package]*packages.Package

	typeGlobalVarToValueMap map[string]string
}

func NewPayloadProcessor(hc context.HandlerContext) *EchoPayloadProcessor {
	return &EchoPayloadProcessor{
		resolvedAssignExpr:      make(map[string]string),
		visitedVar:              make(map[*types.Var]bool),
		typesInfo:               hc.GetPackage().TypesInfo,
		typeNameToGenDeclCache:  hc.GetTypeNameToGenDeclCache(),
		typePackageCache:        hc.GetTypePackageCache(),
		typeGlobalVarToValueMap: hc.GetTypeGlobalVarToValueMap(),
	}
}

func (p *EchoPayloadProcessor) UpdateCacheFromDeclStmt(n *ast.DeclStmt) {
	genDecl, ok := n.Decl.(*ast.GenDecl)
	if !ok {
		return
	}
	// for now, just handle for the 1 assigment param (no tuple)
	valueSpec, ok := genDecl.Specs[0].(*ast.ValueSpec)
	if !ok {
		return
	}

	key := valueSpec.Names[0].Name
	val, ok := valueSpec.Values[0].(*ast.BasicLit)
	if !ok {
		return
	}
	p.resolvedAssignExpr[key] = val.Value
}

func (p *EchoPayloadProcessor) UpdateCacheFromAssignStmt(n *ast.AssignStmt) {
	// check the lhs is 'ident' or 'selectionExpr'
	// isLhsStruct := false
	finalKey := ""
	finalVal := ""
	if len(n.Lhs) != 1 {
		// for now, only handle  for simple assigment, not for tuple assigment
		return
	}
	switch lhs := n.Lhs[0].(type) {
	case *ast.SelectorExpr:
		// make sure the key itself is a 'string'
		obj, ok := p.typesInfo.Uses[lhs.Sel]
		if !ok {
			return
		}
		if obj.Type().String() != "string" {
			return
		}
		finalKey = fmt.Sprintf("%s.%s", lhs.X, lhs.Sel.Name)
	case *ast.Ident:
		switch n.Tok {
		case token.ASSIGN:
			finalKey = lhs.Name
		case token.DEFINE:
			obj, ok := p.typesInfo.Defs[lhs]
			if !ok {
				return
			}
			// Check if the type is string
			switch t := obj.Type().Underlying().(type) {
			case *types.Basic:
				if t.Kind() == types.String {
					finalKey = obj.Name()
					// fmt.Println("This is a string variable:", finalKey)
				}
			case *types.Struct:
				finalKey = obj.Name()
				// fmt.Println("This is a struct variable:", finalKey)
				// isLhsStruct = true
			default:
				return
			}
		}
	default:
		return
	}

	// check the rhs to see the value
	if len(n.Rhs) != 1 {
		// for now, only handle  for simple assigment, not for tuple assigment
		return
	}
	switch rhs := n.Rhs[0].(type) {
	case *ast.SelectorExpr:
		// make sure the key itself is a 'string'.
		// cover for below case:
		// _ = <x>.<selectExpr>,
		// this value already exist in cache, so just retrieve it
		obj, ok := p.typesInfo.Uses[rhs.Sel]
		if !ok {
			return
		}
		if obj.Type().String() != "string" {
			return
		}
		selExp := fmt.Sprintf("%s.%s", rhs.X, rhs.Sel.Name)
		// check in cache
		value, ok := p.resolvedAssignExpr[selExp]
		if ok {
			finalVal = value
		}
	case *ast.Ident:
		// case for _ = somevar
		// the the 'somevar' is a string contains some value
		obj, ok := p.typesInfo.Uses[rhs]
		if !ok {
			return
		}
		if obj.Type().String() != "string" {
			return
		}
		// check the value in cache
		value, ok := p.resolvedAssignExpr[rhs.Name]
		if ok {
			finalVal = value
		}
	case *ast.BasicLit:
		finalVal = rhs.Value
	case *ast.CompositeLit:
		p.fetchAllStringField(rhs, finalKey)
	default:
		return
	}

	if len(finalKey) != 0 && len(finalVal) != 0 {
		p.resolvedAssignExpr[finalKey] = finalVal
	}
}

func (p *EchoPayloadProcessor) fetchAllStringField(n *ast.CompositeLit, finalKey string) {
	if len(n.Elts) == 0 {
		return
	}
	out := make(map[string]string)
	for _, elt := range n.Elts {
		keyValExpr, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := keyValExpr.Key.(*ast.Ident)
		if !ok {
			continue
		}
		// check if it is string or not
		obj, ok := p.typesInfo.Uses[key]
		if !ok {
			continue
		}
		if obj.Type().String() != "string" {
			continue
		}

		val, ok := keyValExpr.Value.(*ast.BasicLit)
		if !ok {
			continue
		}
		out[key.Name] = val.Value
	}
	for k, v := range out {
		combinedKey := fmt.Sprintf("%s.%s", finalKey, k)
		p.resolvedAssignExpr[combinedKey] = v
	}
}

func (p *EchoPayloadProcessor) resolveBind(argExp *ast.Expr) (*model.PayloadInfo, bool) {
	var paramIdent *ast.Ident
	switch exp := (*argExp).(type) {
	case *ast.Ident:
		paramIdent = exp
	case *ast.SelectorExpr:
		// for case like `<c>.Bind(<some-struct>.<Selector>)`,
		// extract the <Selector> first
		paramIdent = exp.Sel
	default:
		return nil, false
	}

	obj, ok := p.typesInfo.Uses[paramIdent]
	if !ok {
		return nil, false
	}
	v, ok := obj.(*types.Var)
	if !ok || p.visitedVar[v] {
		return nil, false
	}

	// Get underlying struct type
	typ := v.Type()
	if ptr, ok := typ.(*types.Pointer); ok {
		typ = ptr.Elem()
	}
	named, ok := typ.(*types.Named)
	if !ok {
		return nil, false
	}
	typeName := named.Obj() // *types.TypeName
	decl, ok := p.typeNameToGenDeclCache[typeName]
	if !ok {
		return nil, false
	}
	reqData := new(model.PayloadInfo)
	pType := typeName.Pkg()
	fields := p.populateStructFields(pType, decl)
	reqData.FieldLists = fields
	p.visitedVar[v] = true
	return reqData, true
}

// PopulateStructFields
// Only for the `Bind` BindMethod
func (p *EchoPayloadProcessor) populateStructFields(pType *types.Package, structDecl *ast.GenDecl) []*model.StructField {
	result := []*model.StructField{}
	pkg, ok := p.typePackageCache[pType]
	if !ok {
		log.Println("no pkg found", pType)
		return nil
	}
	ast.Inspect(structDecl, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}
		for _, field := range structType.Fields.List {
			// check the var type
			isPointer := false
			tv, ok := pkg.TypesInfo.Types[field.Type]
			if !ok {
				continue
			}

			typ := tv.Type
			if ptr, ok := typ.(*types.Pointer); ok {
				typ = ptr.Elem()
				isPointer = true
			}

			switch t := typ.(type) {
			case *types.Basic:
				newField := &model.StructField{
					IsPointer: isPointer,
					Name:      field.Names[0].Name,
					VarType:   t.Name(),
				}
				if field.Tag != nil && field.Tag.Value != "" {
					newField.Tag = tagparserutil.ParseTag(field.Tag.Value)
				}
				result = append(result, newField)
			case *types.Named:
				typeName := t.Obj()
				nextStructDecl := p.typeNameToGenDeclCache[typeName]
				out := p.populateStructFields(t.Obj().Pkg(), nextStructDecl)
				result = append(result, out...)
			case *types.Struct:
				// fmt.Println("Inline struct with", t.NumFields(), "fields")
			default:
				log.Printf("Unhandled type: %v =>>%T\n", field.Names[0].Name, t)
			}
		}
		return true
	})
	return result
}

func (p *EchoPayloadProcessor) resolveQueryParam(argExp *ast.Expr) (*model.PayloadInfo, bool) {
	if arg, ok := (*argExp).(*ast.BasicLit); ok {
		// c.QueryParam("some-literal")
		return &model.PayloadInfo{BasicLit: arg.Value}, true
	}

	switch arg := (*argExp).(type) {
	case *ast.Ident:
		obj, ok := p.typesInfo.Uses[arg]
		if !ok {
			return nil, false
		}
		reqData := &model.PayloadInfo{}
		switch v := obj.(type) {
		case *types.Var:
			if p.visitedVar[v] {
				return nil, false
			}
			// reqData.Param = v
			p.visitedVar[v] = true
			if basicLit, ok := p.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
			}
			if basicLit, ok := p.resolvedAssignExpr[v.Name()]; ok {
				reqData.BasicLit = basicLit
			}
			return reqData, true
		case *types.Const:
			if basicLit, ok := p.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
				return reqData, true
			}
			if basicLit, ok := p.resolvedAssignExpr[v.Name()]; ok {
				reqData.BasicLit = basicLit
			}
			return reqData, true
		default:
			return nil, false
		}
	case *ast.SelectorExpr:
		obj, ok := p.typesInfo.Uses[arg.Sel]
		if !ok {
			return nil, false
		}
		x, ok := arg.X.(*ast.Ident)
		if !ok {
			return nil, false
		}
		varname := fmt.Sprintf("%s.%s", x.Name, arg.Sel.Name)
		reqData := &model.PayloadInfo{}
		switch v := obj.(type) {
		case *types.Var:
			if basicLit, ok := p.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
			}
			if basicLit, ok := p.resolvedAssignExpr[varname]; ok {
				reqData.BasicLit = basicLit
			}
			return reqData, true
		case *types.Const:
			if basicLit, ok := p.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
				return reqData, true
			}
			if basicLit, ok := p.resolvedAssignExpr[varname]; ok {
				reqData.BasicLit = basicLit
			}
			return reqData, true
		default:
			return nil, false
		}
	default:
		return nil, false
	}
}

func (p *EchoPayloadProcessor) resolveParam(argExp *ast.Expr) (*model.PayloadInfo, bool) {
	if arg, ok := (*argExp).(*ast.BasicLit); ok {
		// c.QueryParam("some-literal")
		return &model.PayloadInfo{BasicLit: arg.Value}, true
	}

	switch arg := (*argExp).(type) {
	case *ast.Ident:
		obj, ok := p.typesInfo.Uses[arg]
		if !ok {
			return nil, false
		}
		reqData := &model.PayloadInfo{}
		switch v := obj.(type) {
		case *types.Var:
			if p.visitedVar[v] {
				return nil, false
			}
			// reqData.Param = v
			p.visitedVar[v] = true
			if basicLit, ok := p.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
			}
			if basicLit, ok := p.resolvedAssignExpr[v.Name()]; ok {
				reqData.BasicLit = basicLit
			}
			return reqData, true
		case *types.Const:
			if basicLit, ok := p.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
				return reqData, true
			}
			if basicLit, ok := p.resolvedAssignExpr[v.Name()]; ok {
				reqData.BasicLit = basicLit
			}
			return reqData, true
		default:
			return nil, false
		}
	case *ast.SelectorExpr:
		obj, ok := p.typesInfo.Uses[arg.Sel]
		if !ok {
			return nil, false
		}
		x, ok := arg.X.(*ast.Ident)
		if !ok {
			return nil, false
		}
		varname := fmt.Sprintf("%s.%s", x.Name, arg.Sel.Name)
		reqData := &model.PayloadInfo{}
		switch v := obj.(type) {
		case *types.Var:
			if basicLit, ok := p.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
			}
			if basicLit, ok := p.resolvedAssignExpr[varname]; ok {
				reqData.BasicLit = basicLit
			}
			return reqData, true
		case *types.Const:
			if basicLit, ok := p.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
				return reqData, true
			}
			if basicLit, ok := p.resolvedAssignExpr[varname]; ok {
				reqData.BasicLit = basicLit
			}
			return reqData, true
		default:
			return nil, false
		}
	default:
		return nil, false
	}
}
func (p *EchoPayloadProcessor) extractPayloadRequest(callExpr *ast.CallExpr) (*model.PayloadInfo, bool) {
	if len(callExpr.Args) != 1 {
		return nil, false
	}

	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, false
	}
	x, ok := selExpr.X.(*ast.Ident)
	if !ok {
		return nil, false
	}
	ident, ok := p.typesInfo.Uses[x]
	if !ok {
		return nil, false
	}

	if ident.Type().String() != framework.ECHO_CONTEXT_TYPE {
		return nil, false
	}
	bindMethod := selExpr.Sel.Name
	var reqData *model.PayloadInfo

	// extract `&req` => `req`
	argExp := callExpr.Args[0]
	if unaryExp, ok := callExpr.Args[0].(*ast.UnaryExpr); ok && unaryExp.Op == token.AND {
		argExp = unaryExp.X
	}
	switch bindMethod {
	case "Bind":
		if reqData, ok = p.resolveBind(&argExp); !ok {
			return nil, false
		}
	case "QueryParam":
		reqData, ok = p.resolveQueryParam(&argExp)
		if !ok {
			return nil, false
		}
	case "Param":
		reqData, ok = p.resolveParam(&argExp)
		if !ok {
			return nil, false
		}
	default:
		return nil, false
	}
	reqData.BindMethod = bindMethod
	return reqData, true
}

func (p *EchoPayloadProcessor) Process(n ast.Node) *model.PayloadInfo {
	switch expr := n.(type) {
	case *ast.AssignStmt:
		p.UpdateCacheFromAssignStmt(expr)
	case *ast.DeclStmt:
		p.UpdateCacheFromDeclStmt(expr)
	case *ast.CallExpr:
		result, ok := p.extractPayloadRequest(expr)
		if !ok {
			return nil
		}
		return result
	}
	return nil
}

func (p *EchoPayloadProcessor) Match(ast.Node) bool {
	return true
}
