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

func TestNewParser_Success(t *testing.T) {
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
	libCode := `
	package pkg

	import "github.com/labstack/echo/v4"

	func DefaultHandler(c echo.Context) error {
		return nil
	}

	`
	err = tmp.AddNewFileInPackage("pkg", "handler_pkg.go", libCode)
	require.NoError(t, err)
	root := tmp.GetTempFile()

	// execute
	parser, err := NewParser(root)
	require.NoError(t, err)
	// assert
	assert.NotNil(t, parser.fset)
	assert.NotNil(t, parser.mainFuncDecl)
	assert.NotNil(t, parser.ctx)
	assert.Equal(t, "main", parser.mainFuncDecl.Name.Name)
	assert.Equal(t, 2, len(parser.pkgs))
}

func TestNewParser_EmptyRoot(t *testing.T) {
	t.Parallel()
	root := ""

	// execute
	parser, err := NewParser(root)

	assert.Equal(t, "root can't be empty", err.Error())
	// assert
	assert.Nil(t, parser)
}

func TestNewParser_WithVendorFileButNoGoFile(t *testing.T) {
	t.Parallel()
	var err error
	tmp, err := testutil.NewTemporaryTestFile(
		t.TempDir(),
		// testutil.IgnoreVendorFile,
	)
	require.NoError(t, err)
	mainCode := `
	just readme
	`

	err = tmp.AddNewFile("main.txt", mainCode)
	require.NoError(t, err)
	libCode := `
	another unrelated file
	`
	err = tmp.AddNewFileInPackage("pkg", "anotherfile.md", libCode)
	require.NoError(t, err)
	root := tmp.GetTempFile()

	// execute
	parser, err := NewParser(root)
	// assert
	assert.Equal(t, "no package found", err.Error())
	assert.Nil(t, parser)
}

func TestNewParser_WithoutVendorFileAndNoGoFile(t *testing.T) {
	t.Parallel()
	var err error
	tmp, err := testutil.NewTemporaryTestFile(
		t.TempDir(),
		testutil.IgnoreVendorFile,
	)
	require.NoError(t, err)
	mainCode := `
	just readme
	`

	err = tmp.AddNewFile("main.txt", mainCode)
	require.NoError(t, err)
	libCode := `
	another unrelated file
	`
	err = tmp.AddNewFileInPackage("pkg", "anotherfile.md", libCode)
	require.NoError(t, err)
	root := tmp.GetTempFile()

	// execute
	parser, err := NewParser(root)
	// assert
	assert.Contains(t, err.Error(), "./...")
	assert.Nil(t, parser)
}
func TestNewParser_WithoutVendorFileWithGoFile(t *testing.T) {
	t.Parallel()
	var err error
	tmp, err := testutil.NewTemporaryTestFile(
		t.TempDir(),
		testutil.IgnoreVendorFile,
	)
	require.NoError(t, err)
	mainCode := `
	package lib

	import (
		"fmt"
		"os"
		"path/filepath"
		"runtime"
	)

	func GetPath() {
		p, _ := os.Getwd()
		fmt.Println("path", p)
	}

	func GetCurrent() {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		fmt.Println(ex)
		exPath := filepath.Dir(ex)
		fmt.Println(exPath)
	}

	func GetRuntimeCaller() {
		_, filename, _, _ := runtime.Caller(0)
		fmt.Println(filename)
	}

	`

	err = tmp.AddNewFile("lib.go", mainCode)
	require.NoError(t, err)

	root := tmp.GetTempFile()

	// execute
	parser, err := NewParser(root)
	// assert
	assert.Contains(t, err.Error(), "does not contain main module")
	assert.Nil(t, parser)
}

func TestNewParser_WithVendorFileAndGoFile(t *testing.T) {
	t.Parallel()
	var err error
	tmp, err := testutil.NewTemporaryTestFile(
		t.TempDir(),
	)
	require.NoError(t, err)
	mainCode := `
	package lib

	import (
		"fmt"
		"os"
		"path/filepath"
		"runtime"
	)

	func GetPath() {
		p, _ := os.Getwd()
		fmt.Println("path", p)
	}

	func GetCurrent() {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		fmt.Println(ex)
		exPath := filepath.Dir(ex)
		fmt.Println(exPath)
	}

	func GetRuntimeCaller() {
		_, filename, _, _ := runtime.Caller(0)
		fmt.Println(filename)
	}

	`

	err = tmp.AddNewFile("lib.go", mainCode)
	require.NoError(t, err)

	root := tmp.GetTempFile()

	// execute
	parser, err := NewParser(root)
	// assert
	assert.Contains(t, err.Error(), "no main file found")
	assert.Nil(t, parser)
}
