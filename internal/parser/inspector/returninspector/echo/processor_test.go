package echo

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ysmnababan/goswaggen/internal/fileutil"
	"github.com/ysmnababan/goswaggen/internal/model"
	"github.com/ysmnababan/goswaggen/internal/parser/helper"
	"github.com/ysmnababan/goswaggen/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestResolveSchemeType(t *testing.T) {
	tests := []struct {
		name        string
		produceType string
		returnType  string
		expected    string
	}{
		// Plain text responses
		{"HTML", "HTML", "string", "{string}"},
		{"HTMLBlob", "HTMLBlob", "string", "{string}"},
		{"String", "String", "string", "{string}"},

		// JSONP
		{"JSONP", "JSONP", "object", "{string}"},
		{"JSONPBlob", "JSONPBlob", "object", "{string}"},

		// JSON / XML blob
		{"JSONBlob", "JSONBlob", "object", "{string}"},
		{"XMLBlob", "XMLBlob", "object", "{string}"},

		// Binary / file
		{"Blob", "Blob", "[]byte", "{file}"},
		{"Stream", "Stream", "[]byte", "{file}"},
		{"File", "File", "[]byte", "{file}"},
		{"Attachment", "Attachment", "[]byte", "{file}"},
		{"Inline", "Inline", "[]byte", "{file}"},

		// NoContent / Redirect
		{"NoContent", "NoContent", "object", ""},
		{"Redirect", "Redirect", "object", ""},

		// JSON / XML with typed return
		{"JSON-string", "JSON", "string", "{string}"},
		{"JSON-int", "JSON", "int", "{integer}"},
		{"JSON-float", "JSON", "float", "{number}"},
		{"JSON-bool", "JSON", "bool", "{boolean}"},
		{"JSON-byte-array", "JSON", "[]byte", "{string}"},
		{"JSON-array", "JSON", "[]MyStruct", "{array}"},
		{"JSON-object", "JSON", "MyStruct", "{object}"},

		{"XML-string", "XML", "string", "{string}"},
		{"XMLPretty-object", "XMLPretty", "MyStruct", "{object}"},
		{"JSONPretty-array", "JSONPretty", "[]OtherStruct", "{array}"},

		// Default fallback
		{"Unknown", "SomethingElse", "OtherStruct", "{object}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveSchemeType(tt.produceType, tt.returnType)
			if got != tt.expected {
				t.Errorf("resolveSchemeType(%q, %q) = %q; want %q",
					tt.produceType, tt.returnType, got, tt.expected)
			}
		})
	}
}
func TestIsErrorIfStmt(t *testing.T) {
	// setup
	const input = `
	package fib

	type S string

	var a, b, c = len(b), S(c), "hello"

	func fib(x int) int {
		var err error
		if err != nil {
			return 10
		}
		if x < 2 {
			return x
		}
		if err != nil {
			return 10
		}
		if err == nil {
			return 10
		}
		return fib(x-1) - fib(x-2)
	}`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", input, 0)
	require.NoError(t, err)
	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	var conf types.Config
	_, err = conf.Check("test", fset, []*ast.File{f}, &info)
	require.NoError(t, err)

	p := EchoReturnProcessor{typesInfo: &info}
	ifStmt := []*ast.IfStmt{}
	ast.Inspect(f, func(n ast.Node) bool {
		if t, ok := n.(*ast.IfStmt); ok {
			ifStmt = append(ifStmt, t)
			return false
		}
		return true
	})
	assert.Equal(t, 4, len(ifStmt))
	errorIfStmtCounter := 0
	// execute
	for _, stmt := range ifStmt {
		if p.isErrorIfStmt(stmt) {
			errorIfStmtCounter++
		}
	}

	// assert
	assert.Equal(t, 2, errorIfStmtCounter)
}

func TestIsFmWorkStandardResponse_AllTrue(t *testing.T) {
	tmp := t.TempDir()
	var err error
	src, err := testutil.GetVendorTestPath()
	require.NoError(t, err)
	err = fileutil.CopyDir(src, tmp)
	require.NoError(t, err)
	mainCode := `
	package main

	import (
		"net/http"
		"os"
		"github.com/labstack/echo/v4"
	)

	func main() {
		e := echo.New()
		e.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		})
		e.Logger.Fatal(e.Start(":1323"))
	}

func somefun(c echo.Context) error {
	return c.HTML(http.StatusOK, "<strong>Hello, World!</strong>")
}
type User struct{
	Name string
	Email string
}
func somefun2(c echo.Context) error {
	u := &User{
		Name:  "Jon",
		Email: "jon@labstack.com",
	}
	return c.JSON(http.StatusOK, u)
}

func somefun3(c echo.Context) error {
	u := &User{
		Name:  "Jon",
		Email: "joe@labstack.com",
	}
	return c.JSONPretty(http.StatusOK, u, "  ")
}

func somefun4(c echo.Context) error {
	encodedJSON := []byte{} // Encoded JSON from external source
	return c.JSONBlob(http.StatusOK, encodedJSON)
}

func somefun5(c echo.Context) error {
	u := &User{
		Name:  "Jon",
		Email: "jon@labstack.com",
	}
	return c.XML(http.StatusOK, u)
}

func somefun6(c echo.Context) error {
	u := &User{
		Name:  "Jon",
		Email: "joe@labstack.com",
	}
	return c.XMLPretty(http.StatusOK, u, "  ")
}

func somefun7(c echo.Context) error {
	encodedXML := []byte{} // Encoded XML from external source
	return c.XMLBlob(http.StatusOK, encodedXML)
}

func somefun8(c echo.Context) error {
	return c.File("<PATH_TO_YOUR_FILE>")
}

func somefun9(c echo.Context) error {
	return c.Attachment("<PATH_TO_YOUR_FILE>", "<ATTACHMENT_NAME>")
}

func somefun10(c echo.Context) error {
	return c.Inline("<PATH_TO_YOUR_FILE>", "another string")
}

func somefun11(c echo.Context) (err error) {
	data := []byte("0306703,0035866,NO_ACTION,06/19/2006, 0086003,UPDATED,06/19/2006")
	return c.Blob(http.StatusOK, "text/csv", data)
}

func somefun12(c echo.Context) error {
	f, _ := os.Open("<PATH_TO_IMAGE>")
	defer f.Close()
	return c.Stream(http.StatusOK, "image/png", f)
}

func somefun13(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func somefun14(c echo.Context) error {
	return c.Redirect(http.StatusMovedPermanently, "<URL>")
}
	`
	err = os.WriteFile(filepath.Join(tmp, "main.go"), []byte(mainCode), 0644)
	require.NoError(t, err)

	// run `go mod tidy`
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	FSET := token.NewFileSet()
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedImports |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
		Dir:  tmp, // relative to where you run `go run`
		Fset: FSET,
		Env:  append(os.Environ(), "GO111MODULE=on", "GOFLAGS=-mod=vendor"),
	}
	pkgs, err := packages.Load(cfg, "./...") // load add the package
	for _, pkg := range pkgs {
		for _, e := range pkg.Errors {
			t.Fatalf("package load error: %v", e)
		}
	}
	require.NoError(t, err)
	retStmt := []*ast.ReturnStmt{}
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				if ret, ok := n.(*ast.ReturnStmt); ok {
					// fmt.Println(ret.Results[0])
					retStmt = append(retStmt, ret)
					return false
				}
				return true // continue walking
			})
		}
	}
	// execute
	trueCount := 0
	p := EchoReturnProcessor{typesInfo: pkgs[0].TypesInfo}
	for _, stmt := range retStmt {
		if p.isFmworkStandardResponse(stmt) {
			trueCount++
		}
	}
	// assert
	assert.Equal(t, 1, len(pkgs))
	assert.Equal(t, 15, len(retStmt))
	assert.Equal(t, 15, trueCount)
}

func TestIsFmWorkStandardResponse_AllFalse(t *testing.T) {
	tmp := t.TempDir()
	var err error
	src, err := testutil.GetVendorTestPath()
	require.NoError(t, err)
	err = fileutil.CopyDir(src, tmp)
	require.NoError(t, err)
	mainCode := `
	package main

	import (
		"github.com/labstack/echo/v4"
	)
	type errorResponse struct{
		Err error
		Message string
	}
	type User struct{
		Name string
		Email string
	}
	func main() {
		e := echo.New()
		e.GET("/", func(c echo.Context) error {
			return nil
		})
		e.Logger.Fatal(e.Start(":1323"))
	}

	func somefun(c echo.Context) error {
		err:= c.Bind(&User{})
		return err
	}
	func somefun2(c echo.Context) error {
		e:=&errorResponse{}
		return e.Err
	}

	func somefun3(c echo.Context) error {
		e:=&errorResponse{}
		return e.Err
	}

	func somefun4(c echo.Context) error {
		return nil 
	}
	`
	err = os.WriteFile(filepath.Join(tmp, "main.go"), []byte(mainCode), 0644)
	require.NoError(t, err)

	// run `go mod tidy`
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))

	FSET := token.NewFileSet()
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedImports |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
		Dir:  tmp, // relative to where you run `go run`
		Fset: FSET,
		Env:  append(os.Environ(), "GO111MODULE=on", "GOFLAGS=-mod=vendor"),
	}
	pkgs, err := packages.Load(cfg, "./...") // load add the package
	for _, pkg := range pkgs {
		for _, e := range pkg.Errors {
			t.Fatalf("package load error: %v", e)
		}
	}
	require.NoError(t, err)
	retStmt := []*ast.ReturnStmt{}
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				if ret, ok := n.(*ast.ReturnStmt); ok {
					// fmt.Println(ret.Results[0])
					retStmt = append(retStmt, ret)
					return false
				}
				return true // continue walking
			})
		}
	}
	// execute
	falseCount := 0
	p := EchoReturnProcessor{typesInfo: pkgs[0].TypesInfo}
	for _, stmt := range retStmt {
		if !p.isFmworkStandardResponse(stmt) {
			falseCount++
		}
	}
	// assert
	assert.Equal(t, 1, len(pkgs))
	assert.Equal(t, 5, len(retStmt))
	assert.Equal(t, 5, falseCount)
}

func TestResolveStatusCode(t *testing.T) {
	processor := &EchoReturnProcessor{}

	tests := []struct {
		name string
		expr ast.Expr
		want int
	}{
		{
			name: "SelectorExpr http.StatusOK",
			expr: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "http"},
				Sel: &ast.Ident{Name: "StatusOK"},
			},
			want: 200,
		},
		{
			name: "SelectorExpr not http",
			expr: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "custom"},
				Sel: &ast.Ident{Name: "StatusBadRequest"},
			},
			want: 500, // Should reject anything not prefixed with "http"
		},
		{
			name: "Ident StatusBadRequest",
			expr: &ast.Ident{Name: "StatusBadRequest"},
			want: 400,
		},
		{
			name: "BasicLit \"StatusTeapot\"",
			expr: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "StatusTeapot",
			},
			want: 418,
		},
		{
			name: "BasicLit using number status code",
			expr: &ast.BasicLit{
				Kind:  token.STRING,
				Value: "200",
			},
			want: 200,
		},
		{
			name: "Unknown Ident",
			expr: &ast.Ident{Name: "SomethingElse"},
			want: 500,
		},
		{
			name: "Unsupported Expr Type",
			expr: &ast.CallExpr{},
			want: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processor.resolveStatusCode(tt.expr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolvePayloadType(t *testing.T) {
	pkg := types.NewPackage("mypkg", "mypkg")

	myStruct := types.NewTypeName(0, pkg, "MyStruct", nil)
	named := types.NewNamed(myStruct, nil, nil)

	processor := &EchoReturnProcessor{
		typesInfo: &types.Info{
			Types: map[ast.Expr]types.TypeAndValue{},
		},
	}

	// Prepare AST expressions
	identExpr := &ast.Ident{Name: "MyStruct"}
	selectorExpr := &ast.SelectorExpr{
		X:   &ast.Ident{Name: "mypkg"},
		Sel: &ast.Ident{Name: "MyStruct"},
	}

	// Add type mappings (both for ident and selector's Sel)
	processor.typesInfo.Types[identExpr] = types.TypeAndValue{Type: named}
	processor.typesInfo.Types[selectorExpr.Sel] = types.TypeAndValue{Type: named}

	tests := []struct {
		name     string
		expr     ast.Expr
		expected string
	}{
		{
			name:     "Ident",
			expr:     identExpr,
			expected: "mypkg.MyStruct",
		},
		{
			name:     "SelectorExpr",
			expr:     selectorExpr,
			expected: "mypkg.MyStruct",
		},
		{
			name:     "Unknown expr",
			expr:     &ast.CallExpr{},
			expected: "",
		},
		{
			name:     "Unmapped Ident",
			expr:     &ast.Ident{Name: "Unknown"},
			expected: "",
		},
		{
			name: "Pointer to Named type",
			expr: &ast.Ident{Name: "PtrStruct"},
		},
	}

	// Add pointer case to typesInfo
	ptrStructIdent := &ast.Ident{Name: "PtrStruct"}
	ptrType := types.NewPointer(named)
	processor.typesInfo.Types[ptrStructIdent] = types.TypeAndValue{Type: ptrType}
	tests[4].expr = ptrStructIdent
	tests[4].expected = "mypkg.MyStruct"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processor.resolvePayloadType(tt.expr)
			assert.Equal(t, tt.expected, got)
		})
	}
}

// func TestMatch(t *testing.T) {
// 	echopkg := types.NewPackage("github.com/labstack/echo/v4", "echo")
// 	echotypeName := types.NewTypeName(0, echopkg, "Context", nil)
// 	echonamed := types.NewNamed(echotypeName, nil, nil)
// 	echoFun := ast.NewIdent("JSON")

// 	mypkg := types.NewPackage("mypkg", "mypkg")
// 	mytypeName := types.NewTypeName(0, mypkg, "Wrap", nil)
// 	mynamed := types.NewNamed(mytypeName, nil, nil)
// 	myFun := ast.NewIdent("ErrorWrap")

// 	p := &EchoReturnProcessor{
// 		typesInfo: &types.Info{
// 			Uses: make(map[*ast.Ident]types.Object),
// 		},
// 	}
// 	p.typesInfo.Uses[echoFun] = echonamed.Obj()
// 	p.typesInfo.Uses[myFun] = mynamed.Obj()

// 	tests := []struct {
// 		name     string
// 		stmt     ast.Node
// 		expected bool
// 	}{
// 		{
// 			name: "not return stmt",
// 			stmt: &ast.BasicLit{
// 				Value: "some-string",
// 			},
// 			expected: false,
// 		},
// 		{
// 			name: "no result",
// 			stmt: &ast.ReturnStmt{
// 				Results: []ast.Expr{},
// 			},
// 			expected: false,
// 		},
// 		{
// 			name: "len result > 1",
// 			stmt: &ast.ReturnStmt{
// 				Results: []ast.Expr{&ast.BasicLit{}, &ast.BasicLit{}},
// 			},
// 			expected: false,
// 		},
// 		{
// 			name: "echo.JSON()",
// 			stmt: &ast.ReturnStmt{
// 				Results: []ast.Expr{
// 					&ast.CallExpr{
// 						Fun: &ast.SelectorExpr{
// 							X:   ast.NewIdent("c"),
// 							Sel: echoFun,
// 						},
// 					},
// 				},
// 			},
// 			expected: true,
// 		},
// 		{
// 			name: "mypkg.ErrorWrap()",
// 			stmt: &ast.ReturnStmt{
// 				Results: []ast.Expr{
// 					&ast.CallExpr{
// 						Fun: &ast.SelectorExpr{
// 							X:   ast.NewIdent("c"),
// 							Sel: myFun,
// 						},
// 					},
// 				},
// 			},
// 			expected: false,
// 		},
// 		{
// 			name:     "plain text",
// 			stmt:     &ast.BasicLit{Value: "value"},
// 			expected: false,
// 		},
// 		{
// 			name:     "plain error",
// 			stmt:     &ast.Ident{Name: "err"},
// 			expected: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			assert.Equal(t, tt.expected, p.Match(tt.stmt))
// 		})
// 	}
// }

func TestResolveReturnResponse_NotStandardResponse(t *testing.T) {
	respCfg := model.Config{
		DefaultSuccessResponse: "response.APIResponse",
		DefaultFailureResponse: "response.APIResponse",
	}
	p := &EchoReturnProcessor{
		typesInfo: &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Uses:  make(map[*ast.Ident]types.Object),
		},
		cfg: &respCfg,
	}
	pkg := types.NewPackage("myPkg", "myPkg")
	tn := types.NewTypeName(0, pkg, "Wrap", nil)
	x := ast.NewIdent("response")
	fun := &ast.SelectorExpr{
		X:   x,
		Sel: ast.NewIdent("Wrap"),
	}
	p.typesInfo.Uses[x] = tn
	tests := []struct {
		name     string
		isError  bool
		retStmt  *ast.ReturnStmt
		expected model.ReturnResponse
	}{
		{
			name:    "default error",
			isError: true,
			retStmt: &ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CallExpr{
						Fun: fun,
					},
				},
			},
			expected: model.ReturnResponse{
				ReturnDataType: "response.APIResponse",
				StatusCode:     500,
				IsSuccess:      false,
				ProduceType:    "json",
			},
		},
		{
			name:    "default success",
			isError: false,
			retStmt: &ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CallExpr{
						Fun: fun,
					},
				},
			},
			expected: model.ReturnResponse{
				ReturnDataType: "response.APIResponse",
				StatusCode:     200,
				IsSuccess:      true,
				ProduceType:    "json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.resolveReturnResponse(tt.retStmt, tt.isError)
			assert.Equal(t, tt.expected.ProduceType, got.ProduceType)
			assert.Equal(t, tt.expected.IsSuccess, got.IsSuccess)
			assert.Equal(t, tt.expected.StatusCode, got.StatusCode)
			assert.Equal(t, tt.expected.ReturnDataType, got.ReturnDataType)
		})
	}
}

func TestResolveReturnResponse_StandardResponse(t *testing.T) {
	respCfg := model.Config{
		DefaultSuccessResponse: "response.APIResponse",
		DefaultFailureResponse: "response.APIResponse",
	}
	p := &EchoReturnProcessor{
		typesInfo: &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Uses:  make(map[*ast.Ident]types.Object),
		},
		cfg: &respCfg,
	}
	echopkg := types.NewPackage("github.com/labstack/echo/v4", "echo")
	echotypeName := types.NewTypeName(0, echopkg, "Context", nil)
	echonamed := types.NewNamed(echotypeName, nil, nil)
	x := ast.NewIdent("c")

	jsonFun := &ast.SelectorExpr{
		X:   x,
		Sel: ast.NewIdent("JSON"),
	}
	stringFun := &ast.SelectorExpr{
		X:   x,
		Sel: ast.NewIdent("String"),
	}
	p.typesInfo.Uses[x] = echonamed.Obj()

	pkg := types.NewPackage("myPkg", "myPkg")
	retStruct := types.NewTypeName(0, pkg, "User", nil)
	named := types.NewNamed(retStruct, nil, nil)
	returnSelectorExpr := &ast.SelectorExpr{
		X:   ast.NewIdent("myPkg"),
		Sel: ast.NewIdent("User"),
	}
	badReqParam := &ast.SelectorExpr{
		X:   ast.NewIdent("http"),
		Sel: ast.NewIdent("StatusBadRequest"),
	}
	statusOkParam := &ast.SelectorExpr{
		X:   ast.NewIdent("http"),
		Sel: ast.NewIdent("StatusOK"),
	}
	p.typesInfo.Types[returnSelectorExpr.Sel] = types.TypeAndValue{Type: named}
	tests := []struct {
		name     string
		isError  bool
		retStmt  *ast.ReturnStmt
		expected model.ReturnResponse
	}{
		{
			name:    "JSON return false",
			isError: true,
			retStmt: &ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CallExpr{
						Args: []ast.Expr{
							badReqParam,
							returnSelectorExpr,
						},
						Fun: jsonFun,
					},
				},
			},
			expected: model.ReturnResponse{
				ReturnDataType: "myPkg.User",
				StatusCode:     400,
				IsSuccess:      false,
				ProduceType:    "json",
			},
		},
		{
			name:    "JSON Success",
			isError: true,
			retStmt: &ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CallExpr{
						Args: []ast.Expr{
							statusOkParam,
							returnSelectorExpr,
						},
						Fun: jsonFun,
					},
				},
			},
			expected: model.ReturnResponse{
				ReturnDataType: "myPkg.User",
				StatusCode:     200,
				IsSuccess:      true,
				ProduceType:    "json",
			},
		},
		{
			name:    "String Success",
			isError: true,
			retStmt: &ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CallExpr{
						Args: []ast.Expr{
							statusOkParam,
							&ast.BasicLit{Value: "output", Kind: token.STRING},
						},
						Fun: stringFun,
					},
				},
			},
			expected: model.ReturnResponse{
				ReturnDataType: "string",
				StatusCode:     200,
				IsSuccess:      true,
				ProduceType:    "plain",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.resolveReturnResponse(tt.retStmt, tt.isError)
			assert.Equal(t, tt.expected.ProduceType, got.ProduceType)
			assert.Equal(t, tt.expected.IsSuccess, got.IsSuccess)
			assert.Equal(t, tt.expected.StatusCode, got.StatusCode)
			assert.Equal(t, tt.expected.ReturnDataType, got.ReturnDataType)
		})
	}
}

func TestProcess_NonStandardResponse(t *testing.T) {
	tmp, err := testutil.NewTemporaryTestFile(
		t.TempDir(),
		testutil.WithEchoAPIResponsePackage,
	)
	require.NoError(t, err)
	mainCode := `
	package main

	import (
		"fmt"
		"net/http"

		"github.com/labstack/echo/v4"

		"basicapi/response"
	)

	type UserLoginRequest struct {
	}

	func main() {
		e := echo.New()
		e.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		})
		e.GET("/test", Login)
		e.Logger.Fatal(e.Start(":1323"))
	}

	func testLogin() (string, error) {
		return "", nil
	}

	func Login(c echo.Context) error {
		req := &UserLoginRequest{}
		err := c.Bind(req)
		if err != nil {
			return response.Wrap(response.ErrUnprocessableEntity, fmt.Errorf("binding error: %w", err))
		}

		err = c.Validate(req)
		if err != nil {
			return response.Wrap(response.ErrValidation, fmt.Errorf("error validation: %w", err))
		}

		res, err := testLogin()
		if err != nil {
			return err
		}
		// return c.JSON(400, res)
		return response.WithStatusOKResponse(res, c)
	}
	`
	err = tmp.AddNewFile("main.go", mainCode)
	require.NoError(t, err)

	pkgs, err := tmp.BuildPackages()
	require.NoError(t, err)
	targetFunc, _ := helper.SearchDeclFun(pkgs, "Login", &helper.MAIN_PACKAGE_NAME)
	require.NotNil(t, targetFunc)
	var mainPkg *packages.Package
	for _, pkg := range pkgs {
		if pkg.Name == "main" {
			mainPkg = pkg
		}
	}
	out := []*model.ReturnResponse{}
	respCfg := model.Config{
		DefaultSuccessResponse: "response.APIResponse",
		DefaultFailureResponse: "response.APIResponse",
	}
	returnProcessor := NewReturnProcessor(mainPkg.TypesInfo, &respCfg)
	ast.Inspect(targetFunc, func(n ast.Node) bool {
		if ret := returnProcessor.Process(n); ret != nil {
			out = append(out, ret)
		}
		return true
	})

	assert.Equal(t, 4, len(out))
	for _, o := range out {
		assert.Equal(t, "{object}", o.SchemaType)
		assert.Equal(t, "response.APIResponse", o.ReturnDataType)
		assert.Equal(t, "json", o.ProduceType)
	}
	assert.Equal(t, 500, out[0].StatusCode)
	assert.Equal(t, 500, out[1].StatusCode)
	assert.Equal(t, 500, out[2].StatusCode)
	assert.Equal(t, 200, out[3].StatusCode)
	assert.Equal(t, false, out[0].IsSuccess)
	assert.Equal(t, false, out[1].IsSuccess)
	assert.Equal(t, false, out[2].IsSuccess)
	assert.Equal(t, true, out[3].IsSuccess)
}

func TestProcess_StandardResponse(t *testing.T) {
	tmp, err := testutil.NewTemporaryTestFile(
		t.TempDir(),
		testutil.WithEchoAPIResponsePackage,
	)
	require.NoError(t, err)
	mainCode := `
	package main

	import (
		"basicapi/pkg"
		"net/http"
		"strings"

		"github.com/labstack/echo/v4"
	)

	type UserLoginRequest struct {
		Data int
	}

	type Response struct {
	}

	func main() {
		e := echo.New()
		e.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		})
		e.GET("/test", DummyHandler)
		e.Logger.Fatal(e.Start(":1323"))
	}

	func DummyHandler(c echo.Context) error {
		req := &UserLoginRequest{}
		resp := &pkg.User{}
		_ = c.Bind(req)
		someInt:= 20
		switch req.Data {
		case 1:
			// JSON with string
			return c.JSON(200, "somestring") // -> {string}
		case 2:
			// JSON with integer
			return c.JSON(200, 123) // -> {integer}
		case 3:
			// JSON with number
			return c.JSON(200, 3.14) // -> {number}
		case 4:
			// JSON with boolean
			return c.JSON(200, true) // -> {boolean}
		case 5:
			// JSON with array
			return c.JSON(200, []string{"a", "b"}) // -> {array}
		case 6:
			// JSON with object
			return c.JSON(200, resp) // -> {object}
		case 7:
			// XML with object
			return c.XML(200, struct {
				Title string
			}{"Echo XML"}) // -> {object}
		case 8:
			// XML with array
			return c.XML(200, []int{1, 2, 3}) // -> {array}
		case 9:
			// HTML response
			return c.HTML(200, "<h1>Hello HTML</h1>") // -> {string}
		case 10:
			// Plain string response
			return c.String(200, "hello world") // -> {string}
		case 11:
			// File download
			return c.File("somefile.txt") // -> {file}
		case 12:
			// File attachment
			return c.Attachment("somefile.txt", "download.txt") // -> {file}
		case 13:
			// Inline file
			return c.Inline("somefile.txt", "inline.txt") // -> {file}
		case 14:
			// Blob response
			return c.Blob(200, "application/octet-stream", []byte("rawdata")) // -> {file}
		case 15:
			// Stream response
			return c.Stream(200, "application/octet-stream", strings.NewReader("streamdata")) // -> {file}
		case 16:
			// No content
			return c.NoContent(200) // -> ""
		case 17:
			// Redirect
			return c.Redirect(200, "https://example.com") // -> ""
		case 18:
			return c.JSON(200, someInt) // -> {number}
		case 19:
			return c.JSON(200, UserLoginRequest{}) //
		case 20:
			return c.JSON(200, pkg.User{}) // 
		case 21:
			return c.JSON(200, []UserLoginRequest{}) //
		case 22:
			return c.JSON(200, []pkg.User{}) // 
		default:
			// Default JSON object
			return c.JSON(200, struct {
				Message string
			}{"default"}) // -> {object}
		}
	}
	`
	err = tmp.AddNewFile("main.go", mainCode)
	require.NoError(t, err)
	libCode := `
	package pkg

	type User struct {
	}
`
	err = tmp.AddNewFileInPackage("pkg", "pkg.go", libCode)
	require.NoError(t, err)

	pkgs, err := tmp.BuildPackages()
	require.NoError(t, err)
	targetFunc, _ := helper.SearchDeclFun(pkgs, "DummyHandler", &helper.MAIN_PACKAGE_NAME)
	require.NotNil(t, targetFunc)
	var mainPkg *packages.Package
	for _, pkg := range pkgs {
		if pkg.Name == "main" {
			mainPkg = pkg
		}
	}
	out := []*model.ReturnResponse{}
	respCfg := model.Config{
		DefaultSuccessResponse: "default.Success",
		DefaultFailureResponse: "default.Failure",
	}
	returnProcessor := NewReturnProcessor(mainPkg.TypesInfo, &respCfg)
	ast.Inspect(targetFunc, func(n ast.Node) bool {
		if ret := returnProcessor.Process(n); ret != nil {
			out = append(out, ret)
		}
		return true
	})

	assert.Equal(t, 23, len(out))
	for _, o := range out {
		assert.Equal(t, 200, o.StatusCode)
	}
	// Check expected ProduceType, SchemaType, ReturnDataType for each case 1-22
	expected := []struct {
		ProduceType    string
		SchemaType     string
		ReturnDataType string
	}{
		{"json", "{string}", "string"},                 // 1: JSON with string
		{"json", "{integer}", "int"},                   // 2: JSON with integer
		{"json", "{number}", "float"},                  // 3: JSON with number
		{"json", "{boolean}", "bool"},                  // 4: JSON with boolean
		{"json", "{array}", "[]string"},                // 5: JSON with array
		{"json", "{object}", "pkg.User"},               // 6: JSON with object
		{"xml", "{object}", "___"},                     // 7: XML with object (anonymous struct)
		{"xml", "{array}", "[]int"},                    // 8: XML with array
		{"html", "{string}", "string"},                 // 9: HTML response
		{"plain", "{string}", "string"},                // 10: Plain string response
		{"octet-stream", "{file}", ""},                 // 11: File download
		{"octet-stream", "{file}", ""},                 // 12: File attachment
		{"octet-stream", "{file}", ""},                 // 13: Inline file
		{"octet-stream", "{file}", ""},                 // 14: Blob response
		{"octet-stream", "{file}", ""},                 // 15: Stream response
		{"", "", ""},                                   // 16: No content
		{"", "", ""},                                   // 17: Redirect
		{"json", "{integer}", "int"},                   // 18: JSON with someInt
		{"json", "{object}", "main.UserLoginRequest"},  // 19: JSON with UserLoginRequest{}
		{"json", "{object}", "pkg.User"},               // 20: JSON with pkg.User{}
		{"json", "{array}", "[]main.UserLoginRequest"}, // 21: JSON with []UserLoginRequest{}
		{"json", "{array}", "[]pkg.User"},              // 22: JSON with []pkg.User{}
	}
	for i := 0; i < 22; i++ {
		assert.Equal(t, expected[i].ProduceType, out[i].ProduceType, "ProduceType mismatch at case %d", i+1)
		assert.Equal(t, expected[i].SchemaType, out[i].SchemaType, "SchemaType mismatch at case %d", i+1)
		assert.Equal(t, expected[i].ReturnDataType, out[i].ReturnDataType, "ReturnDataType mismatch at case %d", i+1)
	}
	// Default case (23rd)
	assert.Equal(t, "json", out[22].ProduceType)
	assert.Equal(t, "{object}", out[22].SchemaType)
	assert.Equal(t, "___", out[22].ReturnDataType)
}
