package tracking

import (
	"go/ast"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ysmnababan/goswaggen/internal/parser/context"
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
		"basicapi/lib"
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

func TestFindHandlerRegistration_DirectHandler(t *testing.T) {
	t.Parallel()
	tmp, err := testutil.NewTemporaryTestFile(t.TempDir())
	require.NoError(t, err)
	mainCode := `
	package main

	import (
		"basicapi/pkg"
		"net/http"

		"github.com/labstack/echo/v4"
	)

	func main() {
		e := echo.New()
		e.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		})
		e.GET("/one", defaultHandler)
		e.POST("/two", defaultHandler)
		e.PUT("/three", defaultHandler)
		e.DELETE("/four", pkg.DefaultHandler)
		e.PATCH("/longer/path/:id", pkg.DefaultHandler)
		e.Logger.Fatal(e.Start(":1323"))
	}

	func defaultHandler(c echo.Context) error {
		return nil
	}
	`
	err = tmp.AddNewFile("main.go", mainCode)
	require.NoError(t, err)
	libCode := `
	package pkg

	import "github.com/labstack/echo/v4"

	func DefaultHandler(c echo.Context) error {
		return nil
	}

	`
	err = tmp.AddNewFileInPackage("pkg", "handler_pkg.go", libCode)
	require.NoError(t, err)

	pkgs, err := tmp.BuildPackages()
	require.NoError(t, err)
	mainFuncDecl, _ := SearchDeclFun(pkgs, "main", &MAIN_PACKAGE_NAME)
	require.NotNil(t, mainFuncDecl)

	ctx := context.NewRegistrationContext(pkgs, mainFuncDecl)

	// execute
	handlerRegs := FindHandlerRegistration(ctx)

	// assert
	assert.Equal(t, 5, len(handlerRegs))
	assert.Equal(t, "/one", handlerRegs[0].GetFullPath())
	assert.Equal(t, "/two", handlerRegs[1].GetFullPath())
	assert.Equal(t, "/three", handlerRegs[2].GetFullPath())
	assert.Equal(t, "/four", handlerRegs[3].GetFullPath())
	assert.Equal(t, "/longer/path/:id", handlerRegs[4].GetFullPath())

	assert.Equal(t, "GET", handlerRegs[0].GetMethod())
	assert.Equal(t, "POST", handlerRegs[1].GetMethod())
	assert.Equal(t, "PUT", handlerRegs[2].GetMethod())
	assert.Equal(t, "DELETE", handlerRegs[3].GetMethod())
	assert.Equal(t, "PATCH", handlerRegs[4].GetMethod())

	assert.Equal(t, "main.go", filepath.Base(handlerRegs[0].FilePath))
	assert.Equal(t, "main.go", filepath.Base(handlerRegs[1].FilePath))
	assert.Equal(t, "main.go", filepath.Base(handlerRegs[2].FilePath))
	assert.Equal(t, "handler_pkg.go", filepath.Base(handlerRegs[3].FilePath))
	assert.Equal(t, "handler_pkg.go", filepath.Base(handlerRegs[4].FilePath))

	assert.Equal(t, "defaultHandler", handlerRegs[0].FuncDecl.Name.Name)
	assert.Equal(t, "defaultHandler", handlerRegs[1].FuncDecl.Name.Name)
	assert.Equal(t, "defaultHandler", handlerRegs[2].FuncDecl.Name.Name)
	assert.Equal(t, "DefaultHandler", handlerRegs[3].FuncDecl.Name.Name)
	assert.Equal(t, "DefaultHandler", handlerRegs[4].FuncDecl.Name.Name)
}

func TestFindHandlerRegistration_GroupRegistration(t *testing.T) {
	t.Parallel()
	tmp, err := testutil.NewTemporaryTestFile(t.TempDir())
	require.NoError(t, err)
	mainCode := `
	package main

	import (
		"basicapi/pkg"
		"net/http"

		"github.com/labstack/echo/v4"
	)

	func main() {
		e := echo.New()
		e.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		})
		first_group := e.Group("/g1")
		second_group := first_group.Group("/g2")

		first_group.GET("/one", defaultHandler)
		second_group.PATCH("/two", pkg.DefaultHandler)

		e.Logger.Fatal(e.Start(":1323"))
	}

	func defaultHandler(c echo.Context) error {
		return nil
	}
	`
	err = tmp.AddNewFile("main.go", mainCode)
	require.NoError(t, err)
	libCode := `
	package pkg

	import "github.com/labstack/echo/v4"

	func DefaultHandler(c echo.Context) error {
		return nil
	}

	`
	err = tmp.AddNewFileInPackage("pkg", "handler_pkg.go", libCode)
	require.NoError(t, err)

	pkgs, err := tmp.BuildPackages()
	require.NoError(t, err)
	mainFuncDecl, _ := SearchDeclFun(pkgs, "main", &MAIN_PACKAGE_NAME)
	require.NotNil(t, mainFuncDecl)

	ctx := context.NewRegistrationContext(pkgs, mainFuncDecl)

	// execute
	handlerRegs := FindHandlerRegistration(ctx)

	// assert
	assert.Equal(t, 2, len(handlerRegs))
	assert.Equal(t, "/g1/one", handlerRegs[0].GetFullPath())
	assert.Equal(t, "/g1/g2/two", handlerRegs[1].GetFullPath())

	assert.Equal(t, "GET", handlerRegs[0].GetMethod())
	assert.Equal(t, "PATCH", handlerRegs[1].GetMethod())

	assert.Equal(t, "main.go", filepath.Base(handlerRegs[0].FilePath))
	assert.Equal(t, "handler_pkg.go", filepath.Base(handlerRegs[1].FilePath))

	assert.Equal(t, "defaultHandler", handlerRegs[0].FuncDecl.Name.Name)
	assert.Equal(t, "DefaultHandler", handlerRegs[1].FuncDecl.Name.Name)
}
