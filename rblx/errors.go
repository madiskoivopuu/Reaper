package rblx

import "errors"

const (
	UnknownError int = iota
	BadRequest
	AuthorizationDenied
	TokenValidation
	TooManyRequests
)

type Error struct {
	Err  error
	Type int
}

func StatusCodeToError(statusCode int) *Error {
	switch statusCode {
	case 400:
		return NewCustomError(errors.New("Malformed/bad request"), BadRequest)
	case 401:
		return NewCustomError(errors.New("Authorization denied"), AuthorizationDenied)
	case 403:
		return NewCustomError(errors.New("Token Validation failed"), TokenValidation)
	case 429:
		return NewCustomError(errors.New("Too many requests"), TooManyRequests)
	default:
		return NewCustomError(errors.New("Unknown error"), UnknownError)
	}
}

func NewCustomError(err error, errorType int) *Error {
	return &Error{
		Err:  err,
		Type: errorType,
	}
}

func (e Error) Error() string {
	return e.Err.Error()
}
