package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Meta struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"` //internal error code
	Detail     any    `json:"detail,omitempty"`
}

type APIResponse struct {
	Meta Meta `json:"meta"`
	Data any  `json:"data"`
}

func WithStatusOKResponse(data any, c echo.Context) error {
	resp := APIResponse{
		Meta: Meta{
			Success:    true,
			Message:    "Request processed successfully",
			StatusCode: 20001,
		},
		Data: data,
	}
	return c.JSON(http.StatusOK, resp)
}

func WithStatusCreatedResponse(data any, c echo.Context) error {
	resp := APIResponse{
		Meta: Meta{
			Success:    true,
			Message:    "Request created successfully",
			StatusCode: 20101,
		},
		Data: data,
	}
	return c.JSON(http.StatusCreated, resp)
}
