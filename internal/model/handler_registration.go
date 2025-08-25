package model

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

type HandlerRegistration struct {
	Func      *types.Func   // The resolved handler function as obj
	Call      *ast.CallExpr // The actual call expression
	IsDirect  bool          // True if registered directly on Framework instance
	GroupPath string        // Group path segments, e.g., ["/api", "/v1"]
	BasePath  string        // Combined path, e.g., "/api/v1/resource"
	FromFunc  *types.Func   // The function where this registration happens
	Pkg       *packages.Package
	FuncDecl  *ast.FuncDecl // The implementation of the handler function

	//
	// Request []*RequestData // The requested data can be from body payload, query param or param
	// Returns []*ReturnResponse // All `return` inside body function

	PayloadInfo    []*PayloadInfo
	ReturnResponse []*ReturnResponse
}

func (n *HandlerRegistration) Print() {
	fmt.Println(">>>>>>")
	if n.IsDirect {
		method := n.Call.Fun.(*ast.SelectorExpr)
		pathArg := n.Call.Args[0].(*ast.BasicLit)
		fullpath := `"` + n.BasePath + strings.Trim(pathArg.Value, `"`) + `"`
		fmt.Printf("%s.%s(%s,%s)\n", method.X, method.Sel.Name, fullpath, n.Func.Name())
	} else {
		method := n.Call.Fun.(*ast.Ident)
		fmt.Printf("%s(%v)\n", method.String(), n.Call.Args)
	}
	fmt.Println()
}

func (n *HandlerRegistration) GetFuncName() string {
	return n.Func.Name()
}

func (n *HandlerRegistration) GetFuncNameWithPackage() string {
	return fmt.Sprintf("%s.%s", n.Pkg.Name, n.Func.Name())
}

func (n *HandlerRegistration) GetMethod() string {
	method := n.Call.Fun.(*ast.SelectorExpr)
	return method.Sel.Name
}

func (n *HandlerRegistration) GetFullPath() string {
	pathArg := n.Call.Args[0].(*ast.BasicLit)
	return n.BasePath + strings.Trim(pathArg.Value, `"`)
}

func (n *HandlerRegistration) GetPackageName() string {
	return n.Pkg.Name
}

func (n *HandlerRegistration) GetFrameworkName() string {
	return "echo"
}

func (n *HandlerRegistration) GetPayloadInfos() []*PayloadInfo {
	return n.PayloadInfo
}

func (n *HandlerRegistration) ReturnResponses() []*ReturnResponse {
	return n.ReturnResponse
}

// type StructField struct {
// 	Name      string
// 	VarType   string
// 	Tag       map[string]string
// 	IsPointer bool
// }
// type RequestData struct {
// 	Call      *ast.CallExpr // The actual call expression
// 	Param     *types.Var    // The param type
// 	ParamDecl *ast.GenDecl  // Declaration of the param
// 	BasicLit  string        // for queryparam and param args, e.g. <context>.Param("this")

// 	// Only for the `Bind()` method for
// 	// storing the parameter's package type.
// 	// The actual package can be found by searching
// 	// through all the package again
// 	PkgTypes *types.Package
// 	// Body, QueryParam, Param.
// 	// When it is 'Body', the Fields will be populated
// 	// according to the handler HTTP method.
// 	// If its 'POST/PUT/PATCH' => no need to populate the fields

// 	BindMethod string

// 	// For storing all the field from a struct when
// 	// calling the `Bind()` binding function.
// 	// Depends on the BindMethod and HTTP method
// 	FieldLists []*StructField
// }
