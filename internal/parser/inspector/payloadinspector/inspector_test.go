package payloadinspector

import (
	"fmt"
	"go/ast"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ysmnababan/goswaggen/internal/parser/context"
	"github.com/ysmnababan/goswaggen/internal/parser/helper"
	"github.com/ysmnababan/goswaggen/internal/parser/tracking"
	"github.com/ysmnababan/goswaggen/internal/testutil"
)

func TestProcess_StandardResponse_Bind(t *testing.T) {
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
		"encoding/json"
		"fmt"

		"github.com/labstack/echo/v4"
	)
	type OtherReference struct {
		ID   int
		Code string
	}
	type Personal struct {
		Age   int
		Hobby *string
	}
	type UserLoginRequest struct {
	Name        string            ` + "`json:\"name\" validate:\"required\"`" + `     // basic
	Email       string            ` + "`json:\"email\" validate:\"required\"`" + `    // basic
	Password    string            ` + "`json:\"password\" validate:\"required\"`" + ` // basic
	Birthdate   *string           ` + "`json:\"birthdate\"`" + `                    // pointer to basic
	Personal    Personal          // named struct
	Metadata    map[string]string // map
	Tags        []string          // slice
	Scores      []int             // slice of int
	Reference   *OtherReference   // pointer to named type
	Misc        interface{}       // interface
	Raw         json.RawMessage   // alias for []byte
	Coordinates [2]float64        // array
	Callback    func(int) error   // function
	IsActive    bool              // basic
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

	// execute
	pi := NewPayloadInspector(handlerCtx)
	ast.Inspect(handlers[0].FuncDecl, func(n ast.Node) bool {
		pi.Inspect(n)
		return true
	})
	results := pi.Results

	// assert
	assert.Equal(t, 1, len(results))
	f := results[0].FieldLists
	assert.Equal(t, "Bind", results[0].BindMethod)
	assert.Equal(t, "", results[0].BasicLit)
	assert.Equal(t, "", results[0].ParamTypes)
	assert.Equal(t, 9, len(f))

	assert.Equal(t, "Name", f[0].Name)
	assert.Equal(t, false, f[0].IsPointer)
	assert.NotNil(t, f[0].Tag)
	assert.Equal(t, "string", f[0].VarType)
	assert.Equal(t, "name", f[0].Tag["json"])
	assert.Equal(t, "required", f[0].Tag["validate"])

	assert.Equal(t, "Email", f[1].Name)
	assert.Equal(t, false, f[1].IsPointer)
	assert.NotNil(t, f[1].Tag)
	assert.Equal(t, "string", f[1].VarType)
	assert.Equal(t, "email", f[1].Tag["json"])
	assert.Equal(t, "required", f[1].Tag["validate"])

	assert.Equal(t, "Password", f[2].Name)
	assert.Equal(t, false, f[2].IsPointer)
	assert.NotNil(t, f[2].Tag)
	assert.Equal(t, "string", f[2].VarType)
	assert.Equal(t, "password", f[2].Tag["json"])
	assert.Equal(t, "required", f[2].Tag["validate"])

	assert.Equal(t, "Birthdate", f[3].Name)
	assert.Equal(t, true, f[3].IsPointer)
	assert.NotNil(t, f[3].Tag)
	assert.Equal(t, "string", f[3].VarType)
	assert.Equal(t, "birthdate", f[3].Tag["json"])

	assert.Equal(t, "Age", f[4].Name)
	assert.Equal(t, false, f[4].IsPointer)
	assert.Nil(t, f[4].Tag)
	assert.Equal(t, "int", f[4].VarType)

	assert.Equal(t, "Hobby", f[5].Name)
	assert.Equal(t, true, f[5].IsPointer)
	assert.Nil(t, f[5].Tag)
	assert.Equal(t, "string", f[5].VarType)

	assert.Equal(t, "ID", f[6].Name)
	assert.Equal(t, false, f[6].IsPointer)
	assert.Nil(t, f[6].Tag)
	assert.Equal(t, "int", f[6].VarType)

	assert.Equal(t, "Code", f[7].Name)
	assert.Equal(t, false, f[7].IsPointer)
	assert.Nil(t, f[7].Tag)
	assert.Equal(t, "string", f[7].VarType)

	assert.Equal(t, "IsActive", f[8].Name)
	assert.Equal(t, false, f[8].IsPointer)
	assert.Nil(t, f[8].Tag)
	assert.Equal(t, "bool", f[8].VarType)
}

func TestProcess_StandardResponse_QueryParam(t *testing.T) {
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
	var VARKEY string = "var-key"

	const CONSTKEY string = "const-key"
	type UserCreateRequest struct {
	Name        string            ` + "`json:\"name\" validate:\"required\"`" + `     // basic
	Email       string            ` + "`json:\"email\" validate:\"required\"`" + `    // basic
	Password    string            ` + "`json:\"password\" validate:\"required\"`" + ` // basic
	Birthdate   *string           ` + "`json:\"birthdate\"`" + `                    // pointer to basic
	}
	type Response struct {
	}

	func Login(c echo.Context) error {
		req := &UserCreateRequest{}
		err := c.Bind(req)
		if err != nil {
			return response.Wrap(response.ErrUnprocessableEntity, fmt.Errorf("binding error: %w", err))
		}

		req.Email = "emailz"
		key := "some-key"
		t := c
		t.QueryParam(CONSTKEY)
		t.QueryParam(VARKEY)
		date := t.QueryParam(key)
		t.QueryParam(req.Email)

		// test for collect assigned string
		test1 := "test1"
		c.QueryParam(test1)
		var test2 = "test2"
		c.QueryParam(test2)
		var test3 string = "not test 3"
		test3 = "test3"
		c.QueryParam(test3)
		const test4 = "test4"
		c.QueryParam(test4)

		testReq5 := UserCreateRequest{
			Email: "test5",
		}
		c.QueryParam(testReq5.Email)

		testReq6 := UserCreateRequest{}
		testReq6.Email = "test6"
		c.QueryParam(testReq6.Email)
		test7 := "test7"
		testReq7 := UserCreateRequest{}
		testReq7.Email = test7
		c.QueryParam(testReq7.Email)
		_ = date

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

	// execute
	pi := NewPayloadInspector(handlerCtx)
	ast.Inspect(handlers[0].FuncDecl, func(n ast.Node) bool {
		pi.Inspect(n)
		return true
	})
	results := pi.Results
	// pi.PrintResult()

	// assert
	want := []string{
		"const-key",
		"var-key",
		"some-key",
		"emailz",
		"test1",
		"test2",
		"test3",
		"test4",
		"test5",
		"test6",
		"test7",
	}
	assert.Equal(t, 12, len(results))
	assert.Equal(t, 4, len(results[0].FieldLists))
	for i := 1; i < len(results); i++ {
		assert.Equal(t,
			fmt.Sprintf("%s(\"%s\")\n", "QueryParam", want[i-1]),
			fmt.Sprintf("%s(%s)\n", results[i].BindMethod, results[i].BasicLit),
		)
	}
}
