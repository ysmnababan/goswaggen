package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	t.Run("success search login with package name", func(t *testing.T) {
		err = Generate(
			GeneratePayload{
				root:        root,
				targetFunc:  "pkg.Login",
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

	t.Run("no handler name found", func(t *testing.T) {
		err = Generate(
			GeneratePayload{
				root:        root,
				targetFunc:  "pkg.NoneExistingFunction",
				shouldForce: true,
				config: &config.Config{
					DefaultSuccessResponse: "default.Success",
					DefaultFailureResponse: "default.Failure",
				},
			},
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no handler found")
	})
}

func TestGenerate_Success(t *testing.T) {
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
	root := tmp.GetTempFile()
	require.NoError(t, err)
	t.Run("success", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err = Generate(
			GeneratePayload{
				root:        root,
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
	})

	t.Run("success with security", func(t *testing.T) {
		buf := new(bytes.Buffer)
		err = Generate(
			GeneratePayload{
				root:        root,
				targetFunc:  "Login",
				shouldForce: false,
				config: &config.Config{
					DefaultSuccessResponse: "default.Success",
					DefaultFailureResponse: "default.Failure",
					Security:               "ApiKeyAuth",
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
// @Security ApiKeyAuth
// @Param req body pkg.UserLoginRequest true "change this description"
// @Param id path string true "change this description"
// @Success 200 {object} pkg.Response "success"
// @Failure 500 {object} default.Failure "error"
// @Failure 400 {object} default.Failure "error"
// @Failure 404 {object} default.Failure "error"
// @Router /test/{id} [put]`
		got := buf.String()
		assert.Contains(t, got, want)
	})
}

func normalize(s string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return strings.Join(lines, "\n")
}

func TestGenerate_PreloadComment(t *testing.T) {
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


	// Login is a handler
	// with multiline docs,
	// so make sure all comment is still there
	//
	// @Summary	This is Summary Login
	// @Description	This is the description too
	// @Tags	dontforgetthetag
	// @Accept	json
	// @Produce	json
	// @Param	req body pkg.UserLoginRequest true "change this description"
	// @Param id path string true "change this description"
	// @Success 200 {object} pkg.Response "success"
	// @Failure 500 {object} default.Failure "error"
	// @Failure 400 {object} default.Failure "error"
	// @Failure 404 {object} default.Failure "error"
	// @Router /test/{id} [put]
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
	root := tmp.GetTempFile()
	require.NoError(t, err)
	buf := new(bytes.Buffer)
	err = Generate(
		GeneratePayload{
			root:        root,
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
// Login is a handler
// with multiline docs,
// so make sure all comment is still there
//
// @Summary This is Summary Login
// @Description This is the description too
// @Tags dontforgetthetag
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
	assert.Contains(t, normalize(got), normalize(want))
}

func TestGenerate_PreloadComment_WithoutFuncDocs(t *testing.T) {
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

	// @Summary This is Summary Login
	// @Description This is the description too
	// @Tags dontforgetthetag
	// @Accept json
	// @Produce json
	// @Param req body pkg.UserLoginRequest true "change this description"
	// @Param id path string true "change this description"
	// @Success 200 {object} pkg.Response "success"
	// @Failure 500 {object} default.Failure "error"
	// @Failure 400 {object} default.Failure "error"
	// @Failure 404 {object} default.Failure "error"
	// @Router /test/{id} [put]
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
	root := tmp.GetTempFile()
	require.NoError(t, err)
	buf := new(bytes.Buffer)
	err = Generate(
		GeneratePayload{
			root:        root,
			targetFunc:  "Login",
			shouldForce: false,
			config: &config.Config{
				DefaultSuccessResponse: "default.Success",
				DefaultFailureResponse: "default.Failure",
				Security:               "BearerAuth",
			},
			srcFile: buf,
		},
	)
	require.NoError(t, err)
	want := `
// Login handles PUT /test/:id
//
// @Summary This is Summary Login
// @Description This is the description too
// @Tags dontforgetthetag
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param req body pkg.UserLoginRequest true "change this description"
// @Param id path string true "change this description"
// @Success 200 {object} pkg.Response "success"
// @Failure 500 {object} default.Failure "error"
// @Failure 400 {object} default.Failure "error"
// @Failure 404 {object} default.Failure "error"
// @Router /test/{id} [put]`

	got := buf.String()
	assert.Contains(t, normalize(got), normalize(want))
}

func TestGenerate_MultipleHandlerFound(t *testing.T) {
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
		"basicapi/secondpkg"

		"github.com/labstack/echo/v4"
	)

	func main() {
		e := echo.New()
		e.GET("/", func(c echo.Context) error {
			return c.String(http.StatusOK, "Hello, World!")
		})
		e.PUT("/test/:id", pkg.Login)
		e.PUT("/secondtest/:id", secondpkg.Login)
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
	libCode = `
	package secondpkg

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
	err = tmp.AddNewFileInPackage("secondpkg", "secondpkg.go", libCode)
	require.NoError(t, err)
	buf := new(bytes.Buffer)
	t.Run("duplicate handler, failed", func(t *testing.T) {
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
		require.Error(t, err)
		assert.Contains(t, err.Error(), "multiple handlers found")
		assert.Contains(t, err.Error(), "pkg.Login")
		assert.Contains(t, err.Error(), "secondpkg.Login")
		fmt.Println(err.Error())
	})

	t.Run("duplicate handler, search with package, success", func(t *testing.T) {
		err = Generate(
			GeneratePayload{
				root:        tmp.GetTempFile(),
				targetFunc:  "pkg.Login",
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
	})
}
