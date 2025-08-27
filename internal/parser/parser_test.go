package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ysmnababan/goswaggen/internal/parser/helper"
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
	mainFuncDecl, _ := helper.SearchDeclFun(pkgs, "main", &helper.MAIN_PACKAGE_NAME)
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

func TestGetHandlerByFuncName_NoHandlerFound(t *testing.T) {
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

	root := tmp.GetTempFile()
	parser, err := NewParser(root)
	require.NoError(t, err)

	t.Run("without package name", func(t *testing.T) {
		// execute
		_, err = parser.getHandlerByFuncName("someHandler")

		//assert
		assert.ErrorContains(t, err, "no handler found")
	})

	t.Run("with package name", func(t *testing.T) {
		// execute
		_, err = parser.getHandlerByFuncName("main.someHandler")

		//assert
		assert.ErrorContains(t, err, "no handler found")
	})
}

func TestGetHandlerByFuncName_DuplicateHandler(t *testing.T) {
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
		e.GET("/test", HandlerTest)
		e.GET("/test2", lib.HandlerTest)
		e.Logger.Fatal(e.Start(":1323"))
	}

	func HandlerTest(e echo.Context) error {
		return nil
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

	root := tmp.GetTempFile()
	parser, err := NewParser(root)
	require.NoError(t, err)

	// execute
	_, err = parser.getHandlerByFuncName("HandlerTest")

	//assert
	fmt.Println(err.Error())
	assert.ErrorContains(t, err, "multiple handlers found")
	assert.ErrorContains(t, err, "main.HandlerTest")
	assert.ErrorContains(t, err, "lib.HandlerTest")
}
