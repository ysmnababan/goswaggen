package middleware

import (
	"errors"
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"learn-go/internals/pkg/redisutil"

	"learn-go/internals/pkg/logger"
	"learn-go/internals/utils/response"
	"learn-go/internals/utils/validator"
)

var redis *redisutil.Redis

func Init(e *echo.Echo, fRedis *redisutil.Redis) {
	redis = fRedis

	e.Validator = validator.NewCustomValidator()

	e.Use(
		Recover,
		echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		}),
		logger.WithRequestLogger(),
	)
	e.HTTPErrorHandler = CustomHTTPErrorHandler
}

func Recover(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func(c echo.Context) {
			if r := recover(); r != nil {
				stackTrace := debug.Stack()
				log.Error().Any("error", r).RawJSON("stackTrace", stackTrace).Send()

				c.JSON(500, map[string]any{
					"message": "something went wrong",
				})
			}
		}(c)

		return next(c)
	}
}

func CustomHTTPErrorHandler(err error, c echo.Context) {
	ctx := c.Request().Context()
	logger := zerolog.Ctx(ctx)

	var apiErr *response.APIError
	if errors.As(err, &apiErr) {
		logger.Error().
			Err(err).
			Int("status_code", apiErr.Code).
			Msg(apiErr.Message)

		_ = c.JSON(apiErr.Code, response.APIResponse{
			Meta: response.Meta{
				Success:    false,
				Message:    apiErr.Message,
				StatusCode: apiErr.StatusCode,
				Detail:     apiErr.Detail,
			},
		})
		return
	}

	logger.Error().
		Err(err).
		Str("path", c.Path()).
		Msg("unhandled internal error")
	_ = c.JSON(http.StatusInternalServerError,
		response.APIResponse{
			Meta: response.Meta{
				Success:    false,
				Message:    response.ErrInternalServerError.Message,
				StatusCode: response.ErrInternalServerError.StatusCode,
			},
		})
}
