// apperr/errors.go

package apperr

import "net/http"

type AppError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Raw        error  `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code int, msg string, raw error) *AppError {
	return &AppError{
		StatusCode: code,
		Message:    msg,
		Raw:        raw,
	}
}

// 400 Bad Request
func BadRequest(msg string, raw error) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       "BAD_REQUEST",
		Message:    msg,
		Raw:        raw,
	}
}

// 401 Unauthorized
func Unauthorized(msg string, raw error) *AppError {
	return &AppError{
		StatusCode: http.StatusUnauthorized,
		Code:       "UNAUTHORIZED",
		Message:    msg,
		Raw:        raw,
	}
}

// 403 Forbidden
func Forbidden(msg string, raw error) *AppError {
	return &AppError{
		StatusCode: http.StatusForbidden,
		Code:       "FORBIDDEN",
		Message:    msg,
		Raw:        raw,
	}
}

// 404 Not Found
func NotFound(msg string, raw error) *AppError {
	return &AppError{
		StatusCode: http.StatusNotFound,
		Code:       "NOT_FOUND",
		Message:    msg,
		Raw:        raw,
	}
}

// 409 Conflict
func Conflict(msg string, raw error) *AppError {
	return &AppError{
		StatusCode: http.StatusConflict,
		Code:       "CONFLICT",
		Message:    msg,
		Raw:        raw,
	}
}

// 422 Unprocessable Entity
func UnprocessableEntity(msg string, raw error) *AppError {
	return &AppError{
		StatusCode: http.StatusUnprocessableEntity,
		Code:       "UNPROCESSABLE_ENTITY",
		Message:    msg,
		Raw:        raw,
	}
}

// 429 Too Many Requests
func TooManyRequests(msg string, raw error) *AppError {
	return &AppError{
		StatusCode: http.StatusTooManyRequests,
		Code:       "TOO_MANY_REQUESTS",
		Message:    msg,
		Raw:        raw,
	}
}

// 500 Internal Server Error
func InternalServerError(msg string, raw error) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    msg,
		Raw:        raw,
	}
}

// 503 Service Unavailable
func ServiceUnavailable(msg string, raw error) *AppError {
	return &AppError{
		StatusCode: http.StatusServiceUnavailable,
		Code:       "SERVICE_UNAVAILABLE",
		Message:    msg,
		Raw:        raw,
	}
}
