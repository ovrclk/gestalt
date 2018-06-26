package parse

import (
	"github.com/buger/jsonparser"
	"github.com/ovrclk/gestalt"
)

type Parser interface {
	Parse(data []byte) (Builder, error)
}

type ListParser interface {
	Parse(Context, data []byte) ([]Builder, error)
}

type Builder interface {
	Build() (gestalt.Component, error)
}

type Context interface {
	Lookup(typ string) (Parser, error)
}

type TaskContext interface {
	Context
	Register(typ string, parser Parser) error
}

func NewTasksParser(registry Registry) ListParser {
	return &tasksParser{builtin: registry}
}

type tasksParser struct {
	builtin     Registry
	definitions Registry
}

func (p *tasksParser) Lookup(typ string) (Parser, error) {
	parser, err := p.definitions.Lookup(typ)
	if err == ErrNotFound {
		parser, err = p.builtin.Lookup(typ)
	}
	return parser, err
}

func (p *tasksParser) register(typ string, parser Parser) error {
	return p.definitions.Register(typ, parser)
}

func (p *tasksParser) Parse(Context, data []byte) ([]Builder, error) {
	return nil, nil
}

func NewComponentParser(fn func(DeclSpec) (gestalt.Component, error)) Parser {
	return cmpParser{fn}
}

type cmpParser struct {
	buildfn func(DeclSpec) (gestalt.Component, error)
}

func (p cmpParser) Parse(ctx Context, data []byte) (Builder, error) {
	val, dtype, _, err := jsonparser.Get(data, "name")

	return p
}

func NewComponentBuilder(spec DeclSpec, fn func(DeclSpec) (Builder, error)) Builder {
	return cmpBuilder{fn, spec}
}

type cmpBuilder struct {
	buildfn func(DeclSpec) (gestalt.Component, error)
	spec    DeclSpec
}

func (b cmpBuilder) Build() (gestalt.Component, error) {
	return b.buildfn(b.spec)
}

func resolveParser(ctx Context, data []byte) (Parser, error) {
	val, dtype, _, err := jsonparser.Get(data, "type")

	if err != nil && err != jsonparser.KeyPathNotFoundError {
		return nil, err
	}

	if dtype == jsonparser.String {
	}

	if dtype != jsonparser.NotExist {
	}

}
