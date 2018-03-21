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

type component struct {
	name   string
	action Action
	meta   vars.Meta
}

func NewComponent(name string, action Action) *component {
	return &component{name: name, action: action, meta: vars.NewMeta()}
}

func NoopComponent(name string) *component {
	return NewComponent(name, func(_ Evaluator) error {
		return nil
	})
}

func (c *component) Name() string {
	return c.name
}

func (c *component) IsPassThrough() bool {
	return false
}

func (c *component) WithMeta(m vars.Meta) Component {
	c.meta = c.meta.Merge(m)
	return c
}

func (c *component) Meta() vars.Meta {
	return c.meta
}

func (c *component) Eval(e Evaluator) error {
	if c.action == nil {
		return fmt.Errorf("empty node")
	} else {
		return c.action(e)
	}
}
