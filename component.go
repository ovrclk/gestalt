package gestalt

import (
	"fmt"

	"github.com/ovrclk/gestalt/result"
	"github.com/ovrclk/gestalt/vars"
)

type Component interface {
	Name() string
	IsPassThrough() bool
	Eval(Evaluator) result.Result
	WithMeta(vars.Meta) Component
}

type CompositeComponent interface {
	Component
	Children() []Component
}

type C struct {
	name  string
	build func(Builder) Runable
	meta  vars.Meta
}

func NewComponent(name string, fn func(Builder) Runable) *C {
	return &C{name: name, build: fn, meta: vars.NewMeta()}
}

func NewComponentR(name string, fn Runable) *C {
	return NewComponent(name, func(_ Builder) Runable {
		return fn
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

func (c *C) Eval(e Evaluator) result.Result {
	return c.Build(e.Builder)(e)
}

func (c *C) Build(b Builder) Runable {
	if c.build == nil {
		return func(_ Evaluator) result.Result {
			return result.Error(fmt.Errorf("empty node"))
		}
	}
	return c.build(b)
}
