package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ysmnababan/goswaggen/internal/config"
	"github.com/ysmnababan/goswaggen/internal/testutil"
)

func TestGenerate_WithInjector(t *testing.T) {
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
		e.PUT("/test/:id", pkg.Login)
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

	type Response struct {
		Token string
		Name  string
	}

	type UserLoginRequest struct {
		Email       string            ` + "`json:\"email\" validate:\"required\"`" + `    // basic
		Password    string            ` + "`json:\"password\" validate:\"required\"`" + ` // basic
	}

	func Login(c echo.Context) error {
		req := &UserLoginRequest{}
		res := &Response{}
		err := c.Bind(req)
		if err != nil {
			return response.Wrap(response.ErrUnprocessableEntity, fmt.Errorf("binding error: %w", err))
		}

		err = c.Validate(req)
		if err != nil {
			return response.Wrap(response.ErrValidation, fmt.Errorf("error validation: %w", err))
		}
		id := c.Param("id")
		_ = id
		return c.JSON(200, res)
	}
`
	err = tmp.AddNewFileInPackage("pkg", "pkg.go", libCode)
	require.NoError(t, err)
	root := tmp.GetTempFile()

	// execute
	t.Run("success search login", func(t *testing.T) {

		err = Generate(
			GeneratePayload{
				root:        root,
				targetFunc:  "Login",
				shouldForce: true,
				config: &config.Config{
					DefaultSuccessResponse: "default.Success",
					DefaultFailureResponse: "default.Failure",
				},
			},
		)
		require.NoError(t, err)
		// assert
		target := filepath.Join(root, "pkg", "pkg.go")
		fmt.Println(target)
		gotBytes, err := os.ReadFile(target)
		require.NoError(t, err)
		want := `
// @Summary Login
// @Description Login
// @Tags ______
// @Accept json
// @Produce json
// @Param req body pkg.UserLoginRequest true "change this description"
// @Param id path string true "change this description"
// @Success 200 {object} pkg.Response "success"
// @Failure 500 {object} default.Failure "error"
// @Failure 400 {object} default.Failure "error"
// @Failure 404 {object} default.Failure "error"
// @Router /test/{id} [put]
`
		got := string(gotBytes)
		assert.Contains(t, got, want)
		fmt.Println(got)
	})
}

func TestGenerate(t *testing.T) {
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
		e.PUT("/test/:id", pkg.Login)
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

	type Response struct {
		Token string
		Name  string
	}

	type UserLoginRequest struct {
		Email       string            ` + "`json:\"email\" validate:\"required\"`" + `    // basic
		Password    string            ` + "`json:\"password\" validate:\"required\"`" + ` // basic
	}

	func Login(c echo.Context) error {
		req := &UserLoginRequest{}
		res := &Response{}
		err := c.Bind(req)
		if err != nil {
			return response.Wrap(response.ErrUnprocessableEntity, fmt.Errorf("binding error: %w", err))
		}

		err = c.Validate(req)
		if err != nil {
			return response.Wrap(response.ErrValidation, fmt.Errorf("error validation: %w", err))
		}
		id := c.Param("id")
		_ = id
		return c.JSON(200, res)
	}
`
	err = tmp.AddNewFileInPackage("pkg", "pkg.go", libCode)
	require.NoError(t, err)
	buf := new(bytes.Buffer)
	err = Generate(
		GeneratePayload{
			root:        tmp.GetTempFile(),
			targetFunc:  "Login",
			shouldForce: false,
			config: &config.Config{
				DefaultSuccessResponse: "default.Success",
				DefaultFailureResponse: "default.Failure",
			},
			srcFile: buf,
		},
	)
	require.NoError(t, err)
	want := `
// @Summary Login
// @Description Login
// @Tags ______
// @Accept json
// @Produce json
// @Param req body pkg.UserLoginRequest true "change this description"
// @Param id path string true "change this description"
// @Success 200 {object} pkg.Response "success"
// @Failure 500 {object} default.Failure "error"
// @Failure 400 {object} default.Failure "error"
// @Failure 404 {object} default.Failure "error"
// @Router /test/{id} [put]`
	got := buf.String()
	assert.Contains(t, got, want)
	fmt.Println(got)
}
