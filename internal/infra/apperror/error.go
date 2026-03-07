package apperror

import (
	"errors"
	"net/http"
)

const (
	CodeInvalidArgument = "invalid_argument"
	CodeNotFound        = "not_found"
	CodeConflict        = "conflict"
	CodeInternal        = "internal"
)

type Error struct {
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Code
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(code, message string) *Error {
	return &Error{Code: code, Message: message}
}

func Wrap(code, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

func StatusCode(err error) int {
	var appErr *Error
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case CodeInvalidArgument:
			return http.StatusBadRequest
		case CodeNotFound:
			return http.StatusNotFound
		case CodeConflict:
			return http.StatusConflict
		default:
			return http.StatusInternalServerError
		}
	}
	return http.StatusInternalServerError
}

func CodeOf(err error) string {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return CodeInternal
}
