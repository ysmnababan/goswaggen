package tracking

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ysmnababan/goswaggen/internal/testutil"
)

func TestResolveHandlerExpr_DirectHandler(t *testing.T) {
	tmp, err := testutil.NewTemporaryTestFile(t.TempDir())
	require.NoError(t, err)
	mainCode := `
	package main

	import (
		"net/http"
		"github.com/labstack/echo/v4"
	)

	func handlerTest(c echo.Context) error{
		return nil 
	}

	func main() {
		e := echo.New()
		e.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		})
		e.GET("/test", handlerTest)
		e.Logger.Fatal(e.Start(":1323"))
	}
	`
	err = tmp.AddNewFile("main.go", mainCode)
	require.NoError(t, err)
	pkgs, err := tmp.BuildPackages()
	require.NoError(t, err)
	var astIdent *ast.Ident
	ast.Inspect(pkgs[0].Syntax[0], func(n ast.Node) bool {
		if i, ok := n.(*ast.Ident); ok && i.Name == "handlerTest" {
			astIdent = i
			return false
		}
		return true
	})
	typeFunc, ok := resolveHandlerExpr(pkgs[0], astIdent)
	assert.True(t, ok)
	assert.Equal(t, "handlerTest", typeFunc.Name())
}

func TestResolveHandlerExpr_ImportedHandler(t *testing.T) {
	tmp, err := testutil.NewTemporaryTestFile(t.TempDir())
	require.NoError(t, err)
	mainCode := `
	package main

	import (
		"hohoho/lib"
		"net/http"
		"github.com/labstack/echo/v4"
	)

	func main() {
		e := echo.New()
		e.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		})
		e.GET("/test", lib.HandlerTest)
		e.Logger.Fatal(e.Start(":1323"))
	}
	`
	err = tmp.AddNewFile("main.go", mainCode)
	require.NoError(t, err)
	libCode := `
	package lib

	import (
		"github.com/labstack/echo/v4"
	)

	func HandlerTest(c echo.Context) error {
		return nil 
	}
	`
	err = tmp.AddNewFileInPackage("lib", "lib.go", libCode)
	require.NoError(t, err)

	pkgs, err := tmp.BuildPackages()
	require.NoError(t, err)
	var astIdent *ast.SelectorExpr
	var x *ast.Ident
	ast.Inspect(pkgs[1].Syntax[0], func(n ast.Node) bool {
		if i, ok := n.(*ast.SelectorExpr); ok && i.Sel.Name == "HandlerTest" {
			astIdent = i
			x = i.X.(*ast.Ident)
			return false
		}
		return true
	})
	typeFunc, ok := resolveHandlerExpr(pkgs[1], astIdent)
	assert.True(t, ok)
	assert.Equal(t, "HandlerTest", typeFunc.Name())
	assert.Equal(t, "lib", x.Name)
}
