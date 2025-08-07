package echo

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"parser/fileutil"
	"parser/model"
	"parser/testutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

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

func TestMatch(t *testing.T) {
	echopkg := types.NewPackage("github.com/labstack/echo/v4", "echo")
	echotypeName := types.NewTypeName(0, echopkg, "Context", nil)
	echonamed := types.NewNamed(echotypeName, nil, nil)
	echoFun := ast.NewIdent("JSON")

	mypkg := types.NewPackage("mypkg", "mypkg")
	mytypeName := types.NewTypeName(0, mypkg, "Wrap", nil)
	mynamed := types.NewNamed(mytypeName, nil, nil)
	myFun := ast.NewIdent("ErrorWrap")

	p := &EchoReturnProcessor{
		typesInfo: &types.Info{
			Uses: make(map[*ast.Ident]types.Object),
		},
	}
	p.typesInfo.Uses[echoFun] = echonamed.Obj()
	p.typesInfo.Uses[myFun] = mynamed.Obj()

	tests := []struct {
		name     string
		stmt     ast.Node
		expected bool
	}{
		{
			name: "not return stmt",
			stmt: &ast.BasicLit{
				Value: "some-string",
			},
			expected: false,
		},
		{
			name: "no result",
			stmt: &ast.ReturnStmt{
				Results: []ast.Expr{},
			},
			expected: false,
		},
		{
			name: "len result > 1",
			stmt: &ast.ReturnStmt{
				Results: []ast.Expr{&ast.BasicLit{}, &ast.BasicLit{}},
			},
			expected: false,
		},
		{
			name: "echo.JSON()",
			stmt: &ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("c"),
							Sel: echoFun,
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "mypkg.ErrorWrap()",
			stmt: &ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("c"),
							Sel: myFun,
						},
					},
				},
			},
			expected: false,
		},
		{
			name:     "plain text",
			stmt:     &ast.BasicLit{Value: "value"},
			expected: false,
		},
		{
			name:     "plain error",
			stmt:     &ast.Ident{Name: "err"},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, p.Match(tt.stmt))
		})
	}
}

func TestResolveReturnResponse_NotStandardResponse(t *testing.T) {
	p := &EchoReturnProcessor{
		typesInfo: &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Uses:  make(map[*ast.Ident]types.Object),
		},
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
				StructType: "response.APIResponse",
				StatusCode: 500,
				IsSuccess:  false,
				AcceptType: "json",
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
				StructType: "response.APIResponse",
				StatusCode: 200,
				IsSuccess:  true,
				AcceptType: "json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.resolveReturnResponse(tt.retStmt, tt.isError)
			assert.Equal(t, tt.expected.AcceptType, got.AcceptType)
			assert.Equal(t, tt.expected.IsSuccess, got.IsSuccess)
			assert.Equal(t, tt.expected.StatusCode, got.StatusCode)
			assert.Equal(t, tt.expected.StructType, got.StructType)
		})
	}
}

func TestResolveReturnResponse_StandardResponse(t *testing.T) {
	p := &EchoReturnProcessor{
		typesInfo: &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
			Uses:  make(map[*ast.Ident]types.Object),
		},
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
				StructType: "myPkg.User",
				StatusCode: 400,
				IsSuccess:  false,
				AcceptType: "JSON",
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
				StructType: "myPkg.User",
				StatusCode: 200,
				IsSuccess:  true,
				AcceptType: "JSON",
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
							&ast.BasicLit{Value: "output"},
						},
						Fun: stringFun,
					},
				},
			},
			expected: model.ReturnResponse{
				StructType: "",
				StatusCode: 200,
				IsSuccess:  true,
				AcceptType: "String",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.resolveReturnResponse(tt.retStmt, tt.isError)
			assert.Equal(t, tt.expected.AcceptType, got.AcceptType)
			assert.Equal(t, tt.expected.IsSuccess, got.IsSuccess)
			assert.Equal(t, tt.expected.StatusCode, got.StatusCode)
			assert.Equal(t, tt.expected.StructType, got.StructType)
		})
	}
}
