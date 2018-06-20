package parse

import (
	"fmt"
)

var (
	ErrNotFound = fmt.Errorf("Not Found")
)

type Registry interface {
	Register(typ string, parser Parser) error
	Lookup(typ string) (Parser, error)
}

type registry map[string]Parser

func NewRegistry() Registry {
	return make(registry)
}

func (r registry) Register(typ string, parser Parser) error {
	if _, ok := r[typ]; ok {
		return fmt.Errorf("registry error: %v already exists", name)
	}
	r[typ] = parser
	return nil
}

func (r registry) Lookup(typ string) (Parser, error) {
	parser, ok := r[typ]
	if !ok {
		return parser, ErrNotFound
	}
	return parser, nil
}

type DefinitionSpec struct {
	Anonymous      bool
	AcceptArgs     bool
	AcceptChildren bool
}

type DeclSpec struct {
	Type string
	Name string

	Args []string
	Run  []DeclSpec

	Require []string
	Export  []string
}

// Parse() -> Builder() -> Component()
