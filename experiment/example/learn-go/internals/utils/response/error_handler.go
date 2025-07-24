package response

import (
	"fmt"
	"net/http"
)

type APIError struct {
	Code       int    `json:"code"`             // HTTP status code
	StatusCode int    `json:"status_code"`      // Internal app-specific error code
	Message    string `json:"message"`          // Friendly message
	Detail     any    `json:"detail,omitempty"` //
	Err        error  `json:"-"`                // Internal cause (for logging)
}

var (
	// --- 400 Bad Request ---
	ErrBadRequest       = &APIError{Code: http.StatusBadRequest, StatusCode: 40001, Message: "Bad request"}
	ErrInputLength      = &APIError{Code: http.StatusBadRequest, StatusCode: 40002, Message: "Max length exceeded"}
	ErrInvalidCharacter = &APIError{Code: http.StatusBadRequest, StatusCode: 40003, Message: "Invalid characters"}
	ErrValidation       = &APIError{Code: http.StatusBadRequest, StatusCode: 40004, Message: "Invalid characters"}

	// --- 401 Unauthorized ---
	ErrUnauthorized           = &APIError{Code: http.StatusUnauthorized, StatusCode: 40101, Message: "Unauthorized, please login"}
	ErrInvalidUserCredentials = &APIError{Code: http.StatusUnauthorized, StatusCode: 40102, Message: "Invalid user credentials"}
	ErrInvalidApplicationID   = &APIError{Code: http.StatusUnauthorized, StatusCode: 40103, Message: "Invalid application ID"}
	ErrExpiredAccessToken     = &APIError{Code: http.StatusUnauthorized, StatusCode: 40104, Message: "Access token has expired"}
	ErrInvalidUserAccount     = &APIError{Code: http.StatusUnauthorized, StatusCode: 40105, Message: "Invalid user account"}
	ErrInvalidPublicAuth      = &APIError{Code: http.StatusUnauthorized, StatusCode: 40106, Message: "Invalid public auth"}

	// --- 403 Forbidden ---
	ErrForbidden              = &APIError{Code: http.StatusForbidden, StatusCode: 40301, Message: "Forbidden"}
	ErrAccountNotInWhitelist  = &APIError{Code: http.StatusForbidden, StatusCode: 40302, Message: "User not in whitelist"}
	ErrForbiddenApiPermission = &APIError{Code: http.StatusForbidden, StatusCode: 40304, Message: "You do not have access to this resource"}

	// --- 404 Not Found ---
	ErrNotFound             = &APIError{Code: http.StatusNotFound, StatusCode: 40401, Message: "Data not found"}
	ErrRouteNotFound        = &APIError{Code: http.StatusNotFound, StatusCode: 40402, Message: "Route not found"}
	ErrAccountNotRegistered = &APIError{Code: http.StatusNotFound, StatusCode: 40403, Message: "Account is not registered. Contact administrator."}

	// --- 409 Conflict ---
	ErrDuplicate         = &APIError{Code: http.StatusConflict, StatusCode: 40901, Message: "Value already exists"}
	ErrAccountRegistered = &APIError{Code: http.StatusConflict, StatusCode: 40902, Message: "Account already registered"}
	ErrOrderDetailChange = &APIError{Code: http.StatusConflict, StatusCode: 40903, Message: "Order detail has changed"}

	// --- 422 Unprocessable Entity ---
	ErrUnprocessableEntity = &APIError{Code: http.StatusUnprocessableEntity, StatusCode: 42201, Message: "Invalid parameters or payload"}
	ErrInvalidOTP          = &APIError{Code: http.StatusUnprocessableEntity, StatusCode: 42205, Message: "Invalid OTP"}
	ErrInvalidToken        = &APIError{Code: http.StatusUnprocessableEntity, StatusCode: 42206, Message: "Invalid token"}

	// --- 429 Too Many Requests ---
	ErrForgotPasswordMaxAttempt = &APIError{Code: http.StatusTooManyRequests, StatusCode: 42901, Message: "Maximum password reset requests reached today"}

	// --- 500 Internal Server Errors ---
	ErrInternalServerError = &APIError{Code: http.StatusInternalServerError, StatusCode: 50001, Message: "Something bad happened"}
)

func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func Wrap(baseErr *APIError, err error, detail ...any) *APIError {
	if baseErr == nil {
		return ErrInternalServerError
	}
	Err := &APIError{
		Code:       baseErr.Code,
		StatusCode: baseErr.StatusCode,
		Message:    baseErr.Message,
		Err:        err,
	}
	if len(detail) > 0 {
		Err.Detail = detail[0]
	}
	return Err
}
