package example_feat

import (
	"learn-go/internals/middleware"

	"github.com/labstack/echo/v4"
)

func (h *handler) Route(g *echo.Group) {
	g.GET("", h.GetUsers, middleware.Authentication)
	g.POST("", h.CreateUser)
	g.POST("/auth", h.Login)
}
