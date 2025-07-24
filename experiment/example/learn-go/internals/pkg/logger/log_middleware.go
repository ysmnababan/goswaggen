package logger

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func WithRequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			logger := log.Logger.With().
				Str("method", c.Request().Method).
				Str("path", c.Path()).
				Str("remote_ip", c.RealIP()).
				Str("user_agent", req.UserAgent()).
				Logger()

			ctx := logger.WithContext(c.Request().Context())
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
