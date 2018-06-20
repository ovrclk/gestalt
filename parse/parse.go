package parse

import "github.com/ovrclk/gestalt"

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
	return &tasksParser{registry: registry}
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

func (p *tasksParser) Parse(data []byte) ([]Builder, error) {
}
