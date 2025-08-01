package echo

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
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

func TestIsFmWorkStandardResponse(t *testing.T) {
	tmp := t.TempDir()
	err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte(
		`
	module testproj

	go 1.24.4

	require github.com/labstack/echo/v4 v4.13.4

	require (
		github.com/labstack/gommon v0.4.2 // indirect
		github.com/mattn/go-colorable v0.1.14 // indirect
		github.com/mattn/go-isatty v0.0.20 // indirect
		github.com/valyala/bytebufferpool v1.0.0 // indirect
		github.com/valyala/fasttemplate v1.2.2 // indirect
		golang.org/x/crypto v0.38.0 // indirect
		golang.org/x/net v0.40.0 // indirect
		golang.org/x/sys v0.33.0 // indirect
		golang.org/x/text v0.25.0 // indirect
	)
		`,
	), 0644)
	require.NoError(t, err)
	mainCode := `
	package main

	import (
		"net/http"

		"github.com/labstack/echo/v4"
	)

	func main() {
		e := echo.New()
		e.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		})
		e.Logger.Fatal(e.Start(":1323"))
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
		Env:  append(os.Environ(), "GO111MODULE=on"),
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
					fmt.Println(ret.Results[0])
					retStmt = append(retStmt, ret)
					return false
				}
				return true // continue walking
			})
		}
	}
	// execute
	p := EchoReturnProcessor{typesInfo: pkgs[0].TypesInfo}
	for _, stmt := range retStmt {
		assert.True(t, p.isFmworkStandardResponse(stmt))
	}
	// require
	assert.Equal(t, 1, len(pkgs))
	// assert
	assert.Equal(t, 1, len(retStmt))
}
