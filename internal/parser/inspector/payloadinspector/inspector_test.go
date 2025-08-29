package payloadinspector

import (
	"go/ast"
	"go/types"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ysmnababan/goswaggen/internal/parser/context"
	"github.com/ysmnababan/goswaggen/internal/parser/helper"
	"github.com/ysmnababan/goswaggen/internal/parser/tracking"
	"github.com/ysmnababan/goswaggen/internal/testutil"
)

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

		"github.com/labstack/echo/v4"
	)

	func main() {
		e := echo.New()
		e.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		})
		e.POST("/test", pkg.Login)
		e.Logger.Fatal(e.Start(":1323"))
	}

	`
	err = tmp.AddNewFile("main.go", mainCode)
	require.NoError(t, err)
	libCode := `
	package pkg

	import (
		"basicapi/response"
		"fmt"

		"github.com/labstack/echo/v4"
	)

	type UserLoginRequest struct {
		Age int
		Email string
	}
	type Response struct {
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
		res := Response{}
		return c.JSON(400, res)
	}
`
	err = tmp.AddNewFileInPackage("pkg", "pkg.go", libCode)
	require.NoError(t, err)

	pkgs, err := tmp.BuildPackages()
	require.NoError(t, err)
	mainFuncDecl, _ := helper.SearchDeclFun(pkgs, "main", &helper.MAIN_PACKAGE_NAME)
	require.NotNil(t, mainFuncDecl)
	ctx := context.NewRegistrationContext(pkgs, mainFuncDecl)
	handlers := tracking.FindHandlerRegistration(ctx)
	require.Equal(t, 1, len(handlers))

	handlerCtx := context.HandlerContext{
		RegCtx:             ctx,
		RegisteredHandler:  handlers[0],
		ExistingVarMap:     make(map[*types.Var]bool),
		ResolvedAssignExpr: make(map[string]string),
	}
	pi := NewPayloadInspector(handlerCtx)
	ast.Inspect(handlers[0].FuncDecl, func(n ast.Node) bool {
		pi.Inspect(n)
		return true
	})
	pi.PrintResult()
}
