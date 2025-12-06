package response

import (
	"net/http"

	"elotus_test/server/logger"

	"github.com/labstack/echo/v4"
)

// ErrorCode represents application error codes
type ErrorCode string

const (
	// General errors
	ErrCodeBadRequest          ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden           ErrorCode = "FORBIDDEN"
	ErrCodeNotFound            ErrorCode = "NOT_FOUND"
	ErrCodeConflict            ErrorCode = "CONFLICT"
	ErrCodeInternalServerError ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeValidationFailed    ErrorCode = "VALIDATION_FAILED"

	// Auth errors
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrCodeTokenInvalid       ErrorCode = "TOKEN_INVALID"
	ErrCodeTokenRevoked       ErrorCode = "TOKEN_REVOKED"

	// User errors
	ErrCodeUserExists   ErrorCode = "USER_EXISTS"
	ErrCodeUserNotFound ErrorCode = "USER_NOT_FOUND"

	// File errors
	ErrCodeFileNotFound     ErrorCode = "FILE_NOT_FOUND"
	ErrCodeFileTooLarge     ErrorCode = "FILE_TOO_LARGE"
	ErrCodeInvalidFileType  ErrorCode = "INVALID_FILE_TYPE"
	ErrCodeFileUploadFailed ErrorCode = "FILE_UPLOAD_FAILED"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   ErrorBody `json:"error"`
}

// ErrorBody contains error details
type ErrorBody struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// --- Error Response Helpers ---

// BadRequest returns a 400 Bad Request error response
func BadRequest(c echo.Context, code ErrorCode, message string, details ...interface{}) error {
	logger.Warnf("[%s] Bad Request: %s", code, message)
	return c.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: getDetails(details),
		},
	})
}

// Unauthorized returns a 401 Unauthorized error response
func Unauthorized(c echo.Context, code ErrorCode, message string, details ...interface{}) error {
	logger.Warnf("[%s] Unauthorized: %s", code, message)
	return c.JSON(http.StatusUnauthorized, ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: getDetails(details),
		},
	})
}

// Forbidden returns a 403 Forbidden error response
func Forbidden(c echo.Context, code ErrorCode, message string, details ...interface{}) error {
	logger.Warnf("[%s] Forbidden: %s", code, message)
	return c.JSON(http.StatusForbidden, ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: getDetails(details),
		},
	})
}

// NotFound returns a 404 Not Found error response
func NotFound(c echo.Context, code ErrorCode, message string, details ...interface{}) error {
	logger.Warnf("[%s] Not Found: %s", code, message)
	return c.JSON(http.StatusNotFound, ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: getDetails(details),
		},
	})
}

// Conflict returns a 409 Conflict error response
func Conflict(c echo.Context, code ErrorCode, message string, details ...interface{}) error {
	logger.Warnf("[%s] Conflict: %s", code, message)
	return c.JSON(http.StatusConflict, ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: getDetails(details),
		},
	})
}

// InternalServerError returns a 500 Internal Server Error response
func InternalServerError(c echo.Context, code ErrorCode, message string, err error) error {
	if err != nil {
		logger.ErrorErr(err, message)
	} else {
		logger.Errorf("[%s] Internal Server Error: %s", code, message)
	}
	return c.JSON(http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

// ValidationError returns a 400 Bad Request with validation details
func ValidationError(c echo.Context, message string, details interface{}) error {
	logger.Warnf("[VALIDATION] %s: %v", message, details)
	return c.JSON(http.StatusBadRequest, ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    ErrCodeValidationFailed,
			Message: message,
			Details: details,
		},
	})
}

// --- Success Response Helpers ---

// Success returns a 200 OK success response with data
func Success(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMessage returns a 200 OK success response with message and data
func SuccessWithMessage(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created returns a 201 Created success response
func Created(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusCreated, SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// NoContent returns a 204 No Content response
func NoContent(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// --- Helper Functions ---

func getDetails(details []interface{}) interface{} {
	if len(details) > 0 {
		return details[0]
	}
	return nil
}

// --- Custom Error Type ---

// AppError represents an application error
type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
	Details interface{}
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// NewAppError creates a new AppError
func NewAppError(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewAppErrorWithDetails creates a new AppError with details
func NewAppErrorWithDetails(code ErrorCode, message string, err error, details interface{}) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
		Details: details,
	}
}

// HandleAppError handles AppError and returns appropriate HTTP response
func HandleAppError(c echo.Context, appErr *AppError) error {
	switch appErr.Code {
	case ErrCodeBadRequest, ErrCodeValidationFailed:
		return BadRequest(c, appErr.Code, appErr.Message, appErr.Details)
	case ErrCodeUnauthorized, ErrCodeInvalidCredentials, ErrCodeTokenExpired, ErrCodeTokenInvalid, ErrCodeTokenRevoked:
		return Unauthorized(c, appErr.Code, appErr.Message, appErr.Details)
	case ErrCodeForbidden:
		return Forbidden(c, appErr.Code, appErr.Message, appErr.Details)
	case ErrCodeNotFound, ErrCodeUserNotFound, ErrCodeFileNotFound:
		return NotFound(c, appErr.Code, appErr.Message, appErr.Details)
	case ErrCodeConflict, ErrCodeUserExists:
		return Conflict(c, appErr.Code, appErr.Message, appErr.Details)
	default:
		return InternalServerError(c, appErr.Code, appErr.Message, appErr.Err)
	}
}
