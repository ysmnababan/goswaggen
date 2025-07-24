package server

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"learn-go/config"
	"learn-go/docs"
	"learn-go/internals/app/example_feat"
	"learn-go/internals/factory"

	echoSwagger "github.com/swaggo/echo-swagger"
)

func Init(e *echo.Echo, f *factory.Factory) {
	cfg := config.Get()

	// index
	e.GET("/", func(c echo.Context) error {
		message := fmt.Sprintf("Welcome to %s", cfg.App.Name)
		return c.String(http.StatusOK, message)
	})

	// doc
	if config.Get().EnableSwagger {
		docs.SwaggerInfo.Title = cfg.App.Name
		docs.SwaggerInfo.Host = cfg.App.URL
		docs.SwaggerInfo.Schemes = []string{cfg.App.Schema, "https"}
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	// routes v1
	api := e.Group("/api/v1")
	e.POST("/inside-init", dummyhandler)
	example_feat.NewHandler(f).Route(api.Group("/users"))
}

func dummyhandler(e echo.Context) error {
	return nil
}
