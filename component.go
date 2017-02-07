package gestalt

import (
	"fmt"

	"github.com/ovrclk/gestalt/vars"
)

type Action func(Evaluator) error

type Component interface {
	Name() string
	IsPassThrough() bool
	Eval(Evaluator) error
	WithMeta(vars.Meta) Component
	Meta() vars.Meta
}

type CompositeComponent interface {
	Component
	Children() []Component
}

type C struct {
	name   string
	action Action
	meta   vars.Meta
}

func NewComponent(name string, action Action) *C {
	return &C{name: name, action: action, meta: vars.NewMeta()}
}

func NoopComponent(name string) *C {
	return NewComponent(name, func(_ Evaluator) error {
		return nil
	})
}

func (c *C) Name() string {
	return c.name
}

func (c *C) IsPassThrough() bool {
	return false
}

func (c *C) WithMeta(m vars.Meta) Component {
	c.meta = c.meta.Merge(m)
	return c
}

func (c *C) Meta() vars.Meta {
	return c.meta
}

func (c *C) Eval(e Evaluator) error {
	if c.action == nil {
		return fmt.Errorf("empty node")
	} else {
		return c.action(e)
	}
}
