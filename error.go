package gestalt

import "fmt"

type ErrorWithDetail interface {
	Error() string
	Detail() string
}

type Error interface {
	ErrorWithDetail
	Path() string
}

type evalError struct {
	path    string
	wrapped error
}

func NewError(path string, wrapped error) *evalError {
	return &evalError{path, wrapped}
}

func (e *evalError) Path() string {
	return e.path
}

func (e *evalError) Error() string {
	return fmt.Sprintf("%v: %v", e.path, e.wrapped.Error())
}

func (e *evalError) String() string {
	return e.Error()
}

func (e *evalError) Detail() string {
	if err, ok := e.wrapped.(ErrorWithDetail); ok {
		return err.Detail()
	}
	return ""
}

func (e *evalError) Wrapped() error {
	return e.wrapped
}
