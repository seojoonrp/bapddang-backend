// apperr/errors.go

package apperr

import "net/http"

type AppError struct {
	StatusCode int    `json:"-"`
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
	return New(http.StatusBadRequest, msg, raw)
}

// 401 Unauthorized
func Unauthorized(msg string, raw error) *AppError {
	return New(http.StatusUnauthorized, msg, raw)
}

// 403 Forbidden
func Forbidden(msg string, raw error) *AppError {
	return New(http.StatusForbidden, msg, raw)
}

// 404 Not Found
func NotFound(msg string, raw error) *AppError {
	return New(http.StatusNotFound, msg, raw)
}

// 409 Conflict
func Conflict(msg string, raw error) *AppError {
	return New(http.StatusConflict, msg, raw)
}

// 422 Unprocessable Entity
func UnprocessableEntity(msg string, raw error) *AppError {
	return New(http.StatusUnprocessableEntity, msg, raw)
}

// 429 Too Many Requests
func TooManyRequests(msg string, raw error) *AppError {
	return New(http.StatusTooManyRequests, msg, raw)
}

// 500 Internal Server Error
func InternalServerError(msg string, raw error) *AppError {
	return New(http.StatusInternalServerError, msg, raw)
}

// 503 Service Unavailable
func ServiceUnavailable(msg string, raw error) *AppError {
	return New(http.StatusServiceUnavailable, msg, raw)
}
