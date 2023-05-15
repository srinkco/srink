package apierrors

import "errors"

var (
	ErrAuthTokenInvalid = errors.New("auth token invalid")
	ErrAuthTokenMissing = errors.New("auth token missing")
	ErrEmptyUrl         = errors.New("empty URL parameter")
	ErrInvalidUri       = errors.New("invalid URI for request")
)
