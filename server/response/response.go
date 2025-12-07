package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Total  int  `json:"total,omitempty"`
	Cached bool `json:"cached,omitempty"`
}

const (
	ErrCodeBadRequest      = "BAD_REQUEST"
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeForbidden       = "FORBIDDEN"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeConflict        = "CONFLICT"
	ErrCodeTooManyRequests = "TOO_MANY_REQUESTS"
	ErrCodeInternalError   = "INTERNAL_ERROR"
	ErrCodeValidation      = "VALIDATION_ERROR"
)

func Success(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

func SuccessWithMeta(c echo.Context, data interface{}, meta *Meta) error {
	return c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

func Created(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

func Error(c echo.Context, statusCode int, code, message string) error {
	return c.JSON(statusCode, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	})
}

func BadRequest(c echo.Context, message string) error {
	return Error(c, http.StatusBadRequest, ErrCodeBadRequest, message)
}

func ValidationError(c echo.Context, message string) error {
	return Error(c, http.StatusBadRequest, ErrCodeValidation, message)
}

func Unauthorized(c echo.Context, message string) error {
	return Error(c, http.StatusUnauthorized, ErrCodeUnauthorized, message)
}

func Forbidden(c echo.Context, message string) error {
	return Error(c, http.StatusForbidden, ErrCodeForbidden, message)
}

func NotFound(c echo.Context, message string) error {
	return Error(c, http.StatusNotFound, ErrCodeNotFound, message)
}

func Conflict(c echo.Context, message string) error {
	return Error(c, http.StatusConflict, ErrCodeConflict, message)
}

func TooManyRequests(c echo.Context, message string, retryAfter float64) error {
	return c.JSON(http.StatusTooManyRequests, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    ErrCodeTooManyRequests,
			Message: message,
		},
		Data: map[string]interface{}{
			"retry_after": retryAfter,
		},
	})
}

func InternalError(c echo.Context, message string) error {
	return Error(c, http.StatusInternalServerError, ErrCodeInternalError, message)
}
