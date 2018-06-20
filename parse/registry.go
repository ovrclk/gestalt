package parse

import (
	"fmt"

	"github.com/ovrclk/gestalt"
)

type Registry interface {
	Register(name string, fn ParseFn) error
}

type registry map[string]ParseFn

func NewRegistry() Registry {
	return make(registry)
}

type ParseFn func() (gestalt.Component, error)

func (r registry) Register(name string, fn ParseFn) error {
	if _, ok := r[name]; ok {
		return fmt.Errorf("registry error: %v already exists", name)
	}
	r[name] = fn
	return nil
}

type Parser interface {
	Lookup(name string) (ParseFn, error)
}

type parser struct {
}

func (p parser) Parse(data []byte) error {
	return nil
}
