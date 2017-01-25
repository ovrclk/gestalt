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

type EvalError struct {
	path    string
	wrapped error
}

func NewError(path string, wrapped error) *EvalError {
	return &EvalError{path, wrapped}
}

func (e *EvalError) Path() string {
	return e.path
}

func (e *EvalError) Error() string {
	return fmt.Sprintf("%v: %v", e.path, e.wrapped.Error())
}

func (e *EvalError) String() string {
	return e.Error()
}

func (e *EvalError) Detail() string {
	if err, ok := e.wrapped.(ErrorWithDetail); ok {
		return err.Detail()
	}
	return ""
}

func (e *EvalError) Wrapped() error {
	return e.wrapped
}
