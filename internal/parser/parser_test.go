package parser

import (
	"fmt"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ysmnababan/goswaggen/internal/fileutil"
	"github.com/ysmnababan/goswaggen/internal/testutil"
	"golang.org/x/tools/go/packages"
)

func TestSearchDeclFun_MainExist(t *testing.T) {
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

	utilCode := `
	package main

	func Add(a, b int) int {
		return a + b
	}
	`
	err = os.WriteFile(filepath.Join(tmp, "util.go"), []byte(utilCode), 0644)
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
	fileCount := 0
	for _, pkg := range pkgs {
		for _, e := range pkg.Errors {
			t.Fatalf("package load error: %v", e)
		}
		fileCount += len(pkg.Syntax)
	}
	require.NoError(t, err)

	// execute
	mainFuncDecl, _ := searchDeclFun(pkgs, "main", &MAIN_PACKAGE_NAME)

	// assert
	assert.NotNil(t, mainFuncDecl)
	assert.Equal(t, 1, len(pkgs))
	assert.Equal(t, 2, fileCount)
	assert.Equal(t, "main", mainFuncDecl.Name.String())
}

func TestSearchDeclFun_MainNotExist(t *testing.T) {
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

	func somefunc() {
		e := echo.New()
		e.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		})
		e.Logger.Fatal(e.Start(":1323"))
	}
	`

	err = tmp.AddNewFile("main.go", mainCode)
	require.NoError(t, err)

	utilCode := `
	package main

	func Add(a, b int) int {
		return a + b
	}
	`
	err = tmp.AddNewFile("util.go", utilCode)
	require.NoError(t, err)

	pkgs, err := tmp.BuildPackages()
	require.NoError(t, err)

	// execute
	mainFuncDecl, _ := searchDeclFun(pkgs, "main", &MAIN_PACKAGE_NAME)

	// assert
	assert.Nil(t, mainFuncDecl)
	assert.Equal(t, 1, len(pkgs))
	assert.Equal(t, 2, tmp.FileCount())
}

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
	mainFuncDecl, _ := searchDeclFun(pkgs, "main", &MAIN_PACKAGE_NAME)
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
	fmt.Println(handlers)
	// assert
	assert.Equal(t, 1, len(pkgs))
	assert.Equal(t, 1, tmp.FileCount())
	assert.Equal(t, 2, len(handlers))
	assert.ElementsMatch(t,
		[]string{
			"main.HandlerOne",
			"main.HandlerTwo",
		},
		handlers,
	)
}
