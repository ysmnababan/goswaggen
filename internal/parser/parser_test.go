package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ysmnababan/goswaggen/internal/parser/tracking"
	"github.com/ysmnababan/goswaggen/internal/testutil"
)

func TestGetAllHandlers(t *testing.T) {
	t.Parallel()
	var err error
	tmp, err := testutil.NewTemporaryTestFile(t.TempDir())
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
		e.GET("/one", HandlerOne)
		e.GET("/two", HandlerTwo)
		e.Logger.Fatal(e.Start(":1323"))
	}

	func HandlerOne(c echo.Context) error{
		return nil 
	}

	func HandlerTwo(c echo.Context) error{
		return nil 
	}
	`

	err = tmp.AddNewFile("main.go", mainCode)
	require.NoError(t, err)

	pkgs, err := tmp.BuildPackages()
	require.NoError(t, err)
	mainFuncDecl, _ := tracking.SearchDeclFun(pkgs, "main", &MAIN_PACKAGE_NAME)
	require.NotNil(t, mainFuncDecl)
	require.Equal(t, "main", mainFuncDecl.Name.Name)
	parser := &parser{
		fset:         tmp.GetFileSet(),
		root:         tmp.GetTempFile(),
		pkgs:         pkgs,
		mainFuncDecl: mainFuncDecl,
	}

	// execute
	handlers := parser.GetAllHandlers()
	// assert
	assert.Equal(t, 1, len(pkgs))
	assert.Equal(t, 1, tmp.FileCount())
	assert.Equal(t, 1, len(handlers))

	assert.ElementsMatch(t,
		[]string{
			"HandlerOne",
			"HandlerTwo",
		},
		*handlers["main"],
	)
}
