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

	"golang.org/x/tools/go/packages"
)

type EchoReturnProcessor struct {
	pkg            *packages.Package
	visitedRetStmt map[*ast.ReturnStmt]bool
}

func NewReturnInspector(hc context.HandlerContext) *EchoReturnProcessor {
	return &EchoReturnProcessor{
		pkg:            hc.GetPackage(),
		visitedRetStmt: make(map[*ast.ReturnStmt]bool),
	}
}

func (i *EchoReturnProcessor) isErrorIfStmt(n *ast.IfStmt) bool {
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
	obj, ok := i.pkg.TypesInfo.Uses[xIdent]
	if !ok {
		return false
	}
	return types.Identical(obj.Type(), types.Universe.Lookup("error").Type())
}

// TODO: handle this kind or response
// c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
func (i *EchoReturnProcessor) isFmworkStandardResponse(n *ast.ReturnStmt) bool {
	if len(n.Results) != 1 {
		return false
	}
	callExpr, ok := n.Results[0].(*ast.CallExpr)
	if !ok {
		return false
	}
	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	x, ok := selExpr.X.(*ast.Ident)
	if !ok {
		return false
	}
	obj, ok := i.pkg.TypesInfo.Uses[x]
	if !ok {
		return false
	}
	if obj.Type().String() != framework.ECHO_CONTEXT_TYPE {
		return false
	}
	if _, ok := framework.ECHO_FRAMEWORK_STANDARD_RESPONSE[selExpr.Sel.Name]; !ok {
		return false
	}
	return true
}

func (i *EchoReturnProcessor) resolveStatusCode(n ast.Expr) int {
	out := 500
	var identString string

	switch p := n.(type) {
	case *ast.SelectorExpr:
		x, ok := p.X.(*ast.Ident)
		if !ok {
			return out
		}
		if x.Name != "http" {
			log.Println("status code is not from standard net/http")
			return out
		}
		identString = p.Sel.Name
	case *ast.Ident:
		identString = p.Name
	case *ast.BasicLit:
		identString = p.Value
	default:
		return 500
	}
	if code, ok := framework.HTTP_STATUS_CODE_MAPPING[identString]; ok {
		out = code
	}
	return out
}

func (i *EchoReturnProcessor) resolvePayloadType(n ast.Expr) string {
	var ident *ast.Ident
	switch p := n.(type) {
	case *ast.SelectorExpr:
		x, ok := p.X.(*ast.Ident)
		if !ok {
			return ""
		}
		log.Println("X:", x.Name)
		ident = p.Sel
	case *ast.Ident:
		ident = p
	}
	vn, ok := i.pkg.TypesInfo.Types[ident]
	if !ok {
		return ""
	}
	vType := vn.Type
	if p, ok := vType.(*types.Pointer); ok {
		vType = p.Elem()
	}
	named, ok := vType.(*types.Named)
	if !ok {
		return ""
	}
	obj := named.Obj()
	return fmt.Sprintf("%s.%s", obj.Pkg().Name(), obj.Name())
}

func (i *EchoReturnProcessor) resolveReturnResponse(ret *ast.ReturnStmt, isErrorResponse bool) *model.ReturnResponse {
	result := model.ReturnResponse{
		ReturnStmt: ret,
		IsSuccess:  !isErrorResponse,
	}
	if i.isFmworkStandardResponse(ret) {
		callExpr := ret.Results[0].(*ast.CallExpr)
		selExpr := callExpr.Fun.(*ast.SelectorExpr)
		result.AcceptType = selExpr.Sel.Name
		paramMap := framework.ECHO_FRAMEWORK_STANDARD_RESPONSE[selExpr.Sel.Name]
		if paramMap[0] != 0 {
			result.StatusCode = i.resolveStatusCode(callExpr.Args[paramMap[0]-1])
		}
		if paramMap[1] != 0 {
			result.StructType = i.resolvePayloadType(callExpr.Args[paramMap[1]-1])
		}
		return &result
	}
	result.AcceptType = "json"
	if isErrorResponse {
		result.StatusCode = 500
		result.StructType = "response.APIResponse" // TODO: change this from config
	} else {
		result.StatusCode = 200
		result.StructType = "response.APIResponse" // TODO: change this from config
	}
	return &result
}

func (i *EchoReturnProcessor) Process(in ast.Node) *model.ReturnResponse {
	IsErrorResponse := false
	var retStmt *ast.ReturnStmt
	switch n := in.(type) {
	case *ast.IfStmt:
		// extract the `return` statement inside `ifstmt`
		if !i.isErrorIfStmt(n) {
			return nil
		}
		IsErrorResponse = true
		for _, stmt := range n.Body.List {
			if ret, ok := stmt.(*ast.ReturnStmt); ok {
				retStmt = ret
				break
			}
		}
		if retStmt == nil {
			log.Println("no return statement found inside IfStmt")
			return nil
		}
	case *ast.ReturnStmt:
		// continue
		retStmt = n
	default:
		return nil
	}
	if ok := i.visitedRetStmt[retStmt]; ok {
		return nil
	}
	i.visitedRetStmt[retStmt] = true
	return i.resolveReturnResponse(retStmt, IsErrorResponse)
}

func (i *EchoReturnProcessor) Name() string {
	return "echo return processor"
}

func (i *EchoReturnProcessor) Match(ast.Node) bool {
	return true
}
