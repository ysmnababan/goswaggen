package echo

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"parser/fileutil"
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
