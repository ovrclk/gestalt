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
	name   string
	action Runable
	meta   vars.Meta
}

func NewComponent(name string, action Runable) *C {
	return &C{name: name, action: action, meta: vars.NewMeta()}
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
	if c.action == nil {
		return result.Error(fmt.Errorf("empty node"))
	} else {
		return c.action(e)
	}
}
