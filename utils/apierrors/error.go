package apierrors

import "fmt"

type Error struct {
	method      string
	code        int
	description string
}

func New(method string, code int, desc string) *Error {
	return &Error{method, code, desc}
}

func (e *Error) Error() string {
	return fmt.Sprintf("failed to %s with status code (%d): %s", e.method, e.code, e.description)
}
