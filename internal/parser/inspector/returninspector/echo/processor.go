package echo

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"strings"

	"github.com/ysmnababan/goswaggen/internal/config"
	"github.com/ysmnababan/goswaggen/internal/model"
	"github.com/ysmnababan/goswaggen/internal/parser/framework"
)

type EchoReturnProcessor struct {
	typesInfo      *types.Info
	visitedRetStmt map[*ast.ReturnStmt]bool
	cfg            *config.Config
}

func NewReturnProcessor(ti *types.Info, cfg *config.Config) *EchoReturnProcessor {
	return &EchoReturnProcessor{
		// typesInfo:      hc.GetTypesInfo(),
		typesInfo:      ti,
		visitedRetStmt: make(map[*ast.ReturnStmt]bool),
		cfg:            cfg,
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
	obj, ok := i.typesInfo.Uses[xIdent]
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
	xIdent, ok := selExpr.X.(*ast.Ident)
	if !ok {
		return false
	}
	obj, ok := i.typesInfo.Uses[xIdent]
	if !ok || obj.Type() == nil {
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

func resolveTypeName(typeInfo *types.Info, ident *ast.Ident) string {
	vn, ok := typeInfo.Types[ident]
	if !ok {
		return ""
	}
	vType := vn.Type
	if p, ok := vType.(*types.Pointer); ok {
		vType = p.Elem()
	}
	named, ok := vType.(*types.Named)
	if !ok {
		return vType.String()
	}
	args := named.TypeArgs()
	genericType := ""
	if args != nil { // for handling response with generic type, e.g. Response[User]
		named := args.At(0).(*types.Named)
		genericType = fmt.Sprintf("[%s.%s]", named.Obj().Pkg().Name(), named.Obj().Name())
	}

	typeName := named.Obj()
	return fmt.Sprintf("%s.%s%s", typeName.Pkg().Name(), typeName.Name(), genericType)
}

func (i *EchoReturnProcessor) resolvePayloadType(n ast.Expr) string {
	switch p := n.(type) {
	case *ast.SelectorExpr:
		x, ok := p.X.(*ast.Ident)
		if !ok {
			return ""
		}
		log.Println("X:", x.Name)
		return resolveTypeName(i.typesInfo, p.Sel)
	case *ast.Ident:
		return resolveTypeName(i.typesInfo, p)
	case *ast.BasicLit:
		return strings.ToLower(p.Kind.String())
	case *ast.CompositeLit:
		switch cmpLit := p.Type.(type) {
		case *ast.Ident:
			return resolveTypeName(i.typesInfo, cmpLit)
		case *ast.StructType:
			// TODO: handle this later
			// fmt.Println("here??", cmpLit.Fields)
			return "___" // to
		case *ast.ArrayType:
			ident, ok := cmpLit.Elt.(*ast.Ident)
			if ok {
				return "[]" + resolveTypeName(i.typesInfo, ident)
			}
			selExpr, ok := cmpLit.Elt.(*ast.SelectorExpr)
			if !ok {
				return ""
			}
			x, ok := selExpr.X.(*ast.Ident)
			if !ok {
				return ""
			}
			return "[]" + fmt.Sprintf("%s.%s", x.Name, selExpr.Sel.Name)
		case *ast.SelectorExpr:
			x, ok := cmpLit.X.(*ast.Ident)
			if !ok {
				return ""
			}
			return fmt.Sprintf("%s.%s", x.Name, cmpLit.Sel.String())
		default:
			fmt.Println("OR HERE")
			return ""
		}
	}
	return ""
}

func (i *EchoReturnProcessor) resolveReturnResponse(ret *ast.ReturnStmt, isErrorResponse bool) *model.ReturnResponse {
	result := model.ReturnResponse{
		// ReturnStmt: ret,
	}
	if i.isFmworkStandardResponse(ret) {
		callExpr := ret.Results[0].(*ast.CallExpr)
		selExpr := callExpr.Fun.(*ast.SelectorExpr)
		ptype, ok := framework.ECHO_PRODUCE_MAP[selExpr.Sel.Name]
		if !ok {
			return nil
		}
		result.ProduceType = ptype
		paramMap := framework.ECHO_FRAMEWORK_STANDARD_RESPONSE[selExpr.Sel.Name]
		result.StatusCode = 200 // standard status code
		if paramMap[0] != 0 {
			result.StatusCode = i.resolveStatusCode(callExpr.Args[paramMap[0]-1])
		}
		if paramMap[1] != 0 {
			result.ReturnDataType = i.resolvePayloadType(callExpr.Args[paramMap[1]-1])
		}
		if result.StatusCode/100 == 2 {
			result.IsSuccess = true
		}
		result.SchemaType = resolveSchemeType(selExpr.Sel.Name, result.ReturnDataType)
		return &result
	}
	result.ProduceType = "json"
	result.SchemaType = "{object}"
	if isErrorResponse {
		result.IsSuccess = false
		result.StatusCode = 500
		result.ReturnDataType = i.cfg.DefaultFailureResponse
	} else {
		result.IsSuccess = true
		result.StatusCode = 200
		result.ReturnDataType = i.cfg.DefaultSuccessResponse
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

// func (i *EchoReturnProcessor) Match(n ast.Node) bool {
// 	retStmt, ok := n.(*ast.ReturnStmt)
// 	if !ok {
// 		return false
// 	}
// 	if len(retStmt.Results) != 1 {
// 		return false
// 	}
// 	callExpr, ok := retStmt.Results[0].(*ast.CallExpr)
// 	if !ok {
// 		return false
// 	}
// 	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
// 	if !ok {
// 		return false
// 	}
// 	obj, ok := i.typesInfo.Uses[selExpr.Sel]
// 	if !ok {
// 		return false
// 	}
// 	return obj.Type().String() == framework.ECHO_CONTEXT_TYPE
// }

func resolveSchemeType(produceType, returnType string) string {
	switch produceType {
	case "HTML", "HTMLBlob", "String":
		return "{string}"
	case "JSONP", "JSONPBlob":
		return "{string}"
	case "JSONBlob":
		return "{string}" // raw bytes
	case "XMLBlob":
		return "{string}"
	case "Blob", "Stream", "File", "Attachment", "Inline":
		return "{file}"
	case "NoContent", "Redirect":
		return "" // no schema
	case "XML", "XMLPretty", "JSON", "JSONPretty":
		switch returnType {
		case "string":
			return "{string}"
		case "int":
			return "{integer}"
		case "float":
			return "{number}"
		case "bool":
			return "{boolean}"
		case "[]byte":
			return "{string}" // binary
		default:
			if strings.Contains(returnType, "[]") {
				return "{array}"
			}
			return "{object}"
		}
	default:
		return "{object}"
	}
}
