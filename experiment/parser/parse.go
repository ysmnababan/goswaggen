package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"

	"golang.org/x/tools/go/packages"
)

func TryParseHandler() {
	FSET = token.NewFileSet()
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			// packages.NeedCompiledGoFiles |
			packages.NeedImports |
			// packages.NeedDeps |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
		Dir:  "../example/learn-go/", // relative to where you run `go run`
		Fset: FSET,
	}
	pkgs, err := packages.Load(cfg, "./...") // load add the package
	if err != nil {
		panic(err)
	}
	if len(pkgs) == 0 {
		log.Println("no package found")
		return
	}

	mainPackageName := "main"
	mainFuncDecl, mainFile := SearchDeclFun(pkgs, "main", &mainPackageName)
	if mainFuncDecl == nil {
		log.Println("no main file found")
		return
	}

	fmwork := FindFrameworkImportIdentName(mainFile, "echo")
	if len(fmwork) == 0 {
		return
	}
	fmt.Println("FRAMEWORK import name", fmwork)
	ctx := NewRegistrationContext(pkgs, mainFuncDecl)
	handlerRegs := FindHandlerRegistration(ctx)
	if len(handlerRegs) == 0 {
		fmt.Println("can't find handler registration")
		return
	}
	fmt.Println(len(handlerRegs))
	targetHandler := "CreateUser"
	var handlerFunc *HandlerRegistration
	for _, val := range handlerRegs {
		if val.Func.Name() == targetHandler {
			handlerFunc = val
		}
	}

	if handlerFunc == nil {
		log.Printf("handler with name '%s' was not found\n", targetHandler)
		return
	}
	fmt.Println(handlerFunc.Func.Name())
	fmt.Println(handlerFunc.GetFullPath())
	fmt.Println(handlerFunc.GetMethod())
	handlerCtx := &HandlerContext{
		RegCtx:             ctx,
		RegisteredHandler:  handlerFunc,
		ExistingVarMap:     make(map[*types.Var]bool),
		ResolvedAssignExpr: make(map[string]string),
	}
	payloadReqs := PopulateHandlerRequestPayload(handlerCtx)
	if len(payloadReqs) == 0 {
		log.Println("no requested data ")
		return
	}

	// for _, req := range out {
	// 	param := ""
	// 	if req.Param != nil {
	// 		param = req.Param.String()
	// 	}
	// 	fmt.Printf("method: %s, param: %v, basicLit: %s\n", req.BindMethod, param, req.BasicLit)
	// 	if req.ParamDecl != nil {
	// 		fmt.Println(req.ParamDecl.Specs[0])
	// 	}
	// }
	for _, b := range payloadReqs {
		if b.BindMethod != "Bind" {
			continue
		}
		if b.ParamDecl == nil {
			log.Println("parameter declaration can't be empty", b)
			continue
		}
		fields := PopulateStructFields(ctx, b.PkgTypes, b.ParamDecl)
		for _, val := range fields {
			fmt.Println(*val)
		}
		b.FieldLists = fields
	}
	// CollectAssignedStringValues(handlerFunc)
}

func CollectFromAssignStmt(h *HandlerRegistration, n *ast.AssignStmt, cache map[string]string) {
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
		obj, ok := h.Pkg.TypesInfo.Uses[lhs.Sel]
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
			obj, ok := h.Pkg.TypesInfo.Defs[lhs]
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
		obj, ok := h.Pkg.TypesInfo.Uses[rhs.Sel]
		if !ok {
			return
		}
		if obj.Type().String() != "string" {
			return
		}
		selExp := fmt.Sprintf("%s.%s", rhs.X, rhs.Sel.Name)
		// check in cache
		value, ok := cache[selExp]
		if ok {
			finalVal = value
		}
	case *ast.Ident:
		// case for _ = somevar
		// the the 'somevar' is a string contains some value
		obj, ok := h.Pkg.TypesInfo.Uses[rhs]
		if !ok {
			return
		}
		if obj.Type().String() != "string" {
			return
		}
		// check the value in cache
		value, ok := cache[rhs.Name]
		if ok {
			finalVal = value
		}
	case *ast.BasicLit:
		finalVal = rhs.Value
	case *ast.CompositeLit:
		structField := fetchAllStringField(h.Pkg, rhs)
		for k, v := range structField {
			combinedKey := fmt.Sprintf("%s.%s", finalKey, k)
			cache[combinedKey] = v
		}
	default:
		return
	}

	if len(finalKey) != 0 && len(finalVal) != 0 {
		cache[finalKey] = finalVal
	}
}

func fetchAllStringField(pkg *packages.Package, n *ast.CompositeLit) map[string]string {
	if len(n.Elts) == 0 {
		return nil
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
		obj, ok := pkg.TypesInfo.Uses[key]
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
	return out
}

func CollectFromDeclStmt(h *HandlerRegistration, n *ast.DeclStmt, cache map[string]string) {
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
	cache[key] = val.Value
}

// TODO : handle param for imported constant, or global var
func SearchBindRequest(ctx *HandlerContext, callExpr *ast.CallExpr) (*RequestData, bool) {
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
	ident, ok := ctx.RegisteredHandler.Pkg.TypesInfo.Uses[x]
	if !ok {
		return nil, false
	}

	if ident.Type().String() != ECHO_CONTEXT_TYPE {
		return nil, false
	}
	bindMethod := selExpr.Sel.Name
	reqData := &RequestData{}

	// extract `&req` => `req`
	argExp := callExpr.Args[0]
	if unaryExp, ok := callExpr.Args[0].(*ast.UnaryExpr); ok && unaryExp.Op == token.AND {
		argExp = unaryExp.X
	}
	ctx.BindArgExpr = &argExp
	switch bindMethod {
	case "Bind":
		if reqData, ok = resolveBind(ctx); !ok {
			return nil, false
		}
	case "QueryParam":
		reqData, ok = resolveQueryParam(ctx)
		if !ok {
			return nil, false
		}
	case "Param":
		reqData, ok = resolveParam(ctx)
		if !ok {
			return nil, false
		}
	default:
		return nil, false
	}
	reqData.Call = callExpr
	reqData.BindMethod = bindMethod
	return reqData, true
}

func PopulateHandlerRequestPayload(ctx *HandlerContext) []*RequestData {
	var result []*RequestData
	ast.Inspect(ctx.RegisteredHandler.FuncDecl, func(n ast.Node) bool {
		switch expr := n.(type) {
		case *ast.AssignStmt:
			CollectFromAssignStmt(ctx.RegisteredHandler, expr, ctx.ResolvedAssignExpr)
		case *ast.DeclStmt:
			CollectFromDeclStmt(ctx.RegisteredHandler, expr, ctx.ResolvedAssignExpr)
		case *ast.CallExpr:
			req, ok := SearchBindRequest(ctx, expr)
			if !ok {
				return true
			}
			result = append(result, req)
		}
		return true
	})
	return result
}

func resolveBind(ctx *HandlerContext) (*RequestData, bool) {
	argExp := ctx.BindArgExpr
	h := ctx.RegisteredHandler
	objMap := ctx.ExistingVarMap
	regCtx := ctx.RegCtx
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

	obj, ok := h.Pkg.TypesInfo.Uses[paramIdent]
	if !ok {
		return nil, false
	}
	v, ok := obj.(*types.Var)
	if !ok || objMap[v] {
		return nil, false

	}
	reqData := new(RequestData)
	reqData.Param = v
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
	reqData.PkgTypes = typeName.Pkg()
	if decl, ok := regCtx.typeVarToGenDeclMap[typeName]; ok {
		reqData.ParamDecl = decl
	}
	objMap[v] = true
	return reqData, true
}

func resolveQueryParam(ctx *HandlerContext) (*RequestData, bool) {
	argExp := ctx.BindArgExpr
	h := ctx.RegisteredHandler
	objMap := ctx.ExistingVarMap
	regCtx := ctx.RegCtx

	if arg, ok := (*argExp).(*ast.BasicLit); ok {
		// c.QueryParam("some-literal")
		return &RequestData{BasicLit: arg.Value}, true
	}

	switch arg := (*argExp).(type) {
	case *ast.Ident:
		obj, ok := h.Pkg.TypesInfo.Uses[arg]
		if !ok {
			return nil, false
		}
		reqData := &RequestData{}
		switch v := obj.(type) {
		case *types.Var:
			if objMap[v] {
				return nil, false
			}
			reqData.Param = v
			objMap[v] = true
			if basicLit, ok := regCtx.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
			}
			if basicLit, ok := ctx.ResolvedAssignExpr[v.Name()]; ok {
				reqData.BasicLit = basicLit
			}
			return reqData, true
		case *types.Const:
			if basicLit, ok := regCtx.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
				return reqData, true
			}
			if basicLit, ok := ctx.ResolvedAssignExpr[v.Name()]; ok {
				reqData.BasicLit = basicLit
			}
			return reqData, true
		default:
			return nil, false
		}
	case *ast.SelectorExpr:
		obj, ok := h.Pkg.TypesInfo.Uses[arg.Sel]
		if !ok {
			return nil, false
		}
		x, ok := arg.X.(*ast.Ident)
		if !ok {
			return nil, false
		}
		varname := fmt.Sprintf("%s.%s", x.Name, arg.Sel.Name)
		reqData := &RequestData{}
		switch v := obj.(type) {
		case *types.Var:
			// if objMap[v] {
			// 	return nil, false
			// }
			reqData.Param = v
			// objMap[v] = true
			if basicLit, ok := regCtx.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
			}
			if basicLit, ok := ctx.ResolvedAssignExpr[varname]; ok {
				reqData.BasicLit = basicLit
			}
			return reqData, true
		case *types.Const:
			if basicLit, ok := regCtx.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
				return reqData, true
			}
			if basicLit, ok := ctx.ResolvedAssignExpr[varname]; ok {
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

func resolveParam(ctx *HandlerContext) (*RequestData, bool) {
	argExp := ctx.BindArgExpr
	h := ctx.RegisteredHandler
	objMap := ctx.ExistingVarMap
	regCtx := ctx.RegCtx

	if arg, ok := (*argExp).(*ast.BasicLit); ok {
		// c.Param("some-literal")
		return &RequestData{BasicLit: arg.Value}, true
	}

	switch arg := (*argExp).(type) {
	case *ast.Ident:
		obj, ok := h.Pkg.TypesInfo.Uses[arg]
		if !ok {
			return nil, false
		}
		reqData := &RequestData{}
		switch v := obj.(type) {
		case *types.Var:
			if objMap[v] {
				return nil, false
			}
			reqData.Param = v
			objMap[v] = true
			if basicLit, ok := regCtx.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
			}
			if basicLit, ok := ctx.ResolvedAssignExpr[v.Name()]; ok {
				reqData.BasicLit = basicLit
			}
			return reqData, true
		case *types.Const:
			if basicLit, ok := regCtx.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
				return reqData, true
			}
			if basicLit, ok := ctx.ResolvedAssignExpr[v.Name()]; ok {
				reqData.BasicLit = basicLit
			}
			return reqData, true
		default:
			return nil, false
		}
	case *ast.SelectorExpr:
		obj, ok := h.Pkg.TypesInfo.Uses[arg.Sel]
		if !ok {
			return nil, false
		}
		x, ok := arg.X.(*ast.Ident)
		if !ok {
			return nil, false
		}
		varname := fmt.Sprintf("%s.%s", x.Name, arg.Sel.Name)
		reqData := &RequestData{}
		switch v := obj.(type) {
		case *types.Var:
			// if objMap[v] {
			// 	return nil, false
			// }
			reqData.Param = v
			// objMap[v] = true
			if basicLit, ok := regCtx.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
			}
			if basicLit, ok := ctx.ResolvedAssignExpr[varname]; ok {
				reqData.BasicLit = basicLit
			}
			return reqData, true
		case *types.Const:
			if basicLit, ok := regCtx.typeGlobalVarToValueMap[v.String()]; ok {
				reqData.BasicLit = basicLit
				return reqData, true
			}
			if basicLit, ok := ctx.ResolvedAssignExpr[varname]; ok {
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

// PopulateStructFields
// Only for the `Bind` BindMethod
func PopulateStructFields(ctx *RegistrationContext, pType *types.Package, structDecl *ast.GenDecl) []*StructField {
	cache := ctx.packageMap
	result := []*StructField{}
	pkg, ok := cache[pType]
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
				newField := &StructField{
					IsPointer: isPointer,
					Name:      field.Names[0].Name,
					VarType:   t.Name(),
				}
				if field.Tag != nil && field.Tag.Value != "" {
					newField.Tag = ParseTag(field.Tag.Value)
				}
				result = append(result, newField)
			case *types.Named:
				typeName := t.Obj()
				nextStructDecl := ctx.typeVarToGenDeclMap[typeName]
				out := PopulateStructFields(ctx, t.Obj().Pkg(), nextStructDecl)
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

type Inspector interface {
	Inspect(ast.Node)
}

type ReturnInspector struct {
	FunDecl *ast.FuncDecl
	Pkg     *packages.Package
	Returns []*ReturnResponse
}

func NewReturnInspector(node *ast.FuncDecl, pkg *packages.Package) *ReturnInspector {
	ret := []*ReturnResponse{}
	return &ReturnInspector{
		FunDecl: node,
		Pkg:     pkg,
		Returns: ret,
	}
}

type ReturnResponse struct {
	ReturnStmt *ast.ReturnStmt
	StructType string
	SchemaType string // {object}, string, int, etc
	StatusCode string
	IsSuccess  bool
	AcceptType string //json, xml, string
}

func (i *ReturnInspector) IsErrorIfStmt(n *ast.IfStmt) bool {
	binExp, ok := n.Cond.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	if binExp.Op != token.NEQ {
		return false
	}
	if yIdent, ok := binExp.Y.(*ast.Ident); !ok || yIdent.Name != "nil" {
		return false
	}
	xIdent, ok := binExp.X.(*ast.Ident)
	if !ok {
		return false
	}
	obj, ok := i.Pkg.TypesInfo.Uses[xIdent]
	if !ok {
		return false
	}
	return types.Identical(obj.Type(), types.Universe.Lookup("error").Type())
}

// TODO: handle this kind or response
// c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
func (i *ReturnInspector) IsFmworkStandardResponse(n *ast.ReturnStmt) bool {
	if len(n.Results) != 1 {
		return false
	}
	selExpr, ok := n.Results[0].(*ast.SelectorExpr)
	if !ok {
		return false
	}
	x, ok := selExpr.X.(*ast.Ident)
	if !ok {
		return false
	}
	obj, ok := i.Pkg.TypesInfo.Uses[x]
	if !ok {
		return false
	}
	if obj.Type().String() != ECHO_CONTEXT_TYPE {
		return false
	}
	if _, ok := ECHO_FRAMEWORK_STANDARD_RESPONSE[selExpr.Sel.Name]; !ok {
		return false
	}
	return true
}

func (i *ReturnInspector) ResolveReturnResponse(isErrorResponse bool) *ReturnInspector {
	return nil
}

func (i *ReturnInspector) Inspect(in ast.Node) {
	IsErrorResponse := false
	switch n := in.(type) {
	case *ast.IfStmt:
		// extract the `return` statement inside `ifstmt`
		if !i.IsErrorIfStmt(n) {
			return
		}
		IsErrorResponse = true
		var retStmt *ast.ReturnStmt
		for _, stmt := range n.Body.List {
			if ret, ok := stmt.(*ast.ReturnStmt); ok {
				retStmt = ret
				break
			}
		}
		if retStmt == nil {
			log.Println("no return statement found inside IfStmt")
			return
		}
	case *ast.ReturnStmt:
		// continue
	default:
		return
	}

	i.ResolveReturnResponse(IsErrorResponse)
}
