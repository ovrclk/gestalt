package component

import (
	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/result"
	"github.com/ovrclk/gestalt/vars"
)

type Ensure interface {
	gestalt.Component
	First(gestalt.Component) Ensure
	Run(gestalt.Component) Ensure
	Finally(gestalt.Component) Ensure
}

type EC struct {
	cmp gestalt.Component

	pre   gestalt.Component
	child gestalt.Component
	post  gestalt.Component
}

func NewEnsure(name string) *EC {
	return &EC{
		cmp: gestalt.NewComponent(name, nil),
	}
}

func (c *EC) First(child gestalt.Component) Ensure {
	c.pre = child
	return c
}

func (c *EC) Run(child gestalt.Component) Ensure {
	c.child = child
	return c
}

func (c *EC) Finally(child gestalt.Component) Ensure {
	c.post = child
	return c
}

func (c *EC) Name() string {
	return c.cmp.Name()
}

func (c *EC) Meta() vars.Meta {
	return c.cmp.Meta()
}

func (c *EC) WithMeta(m vars.Meta) gestalt.Component {
	c.cmp.WithMeta(m)
	return c
}

func (c *EC) IsPassThrough() bool {
	return true
}

func (c *EC) Children() []gestalt.Component {
	children := make([]gestalt.Component, 0)
	if c.pre != nil {
		children = append(children, c.pre)
	}
	if c.child != nil {
		children = append(children, c.child)
	}
	if c.post != nil {
		children = append(children, c.post)
	}
	return children
}

func (c *EC) Eval(e gestalt.Evaluator) result.Result {
	if c.pre != nil {
		e.Evaluate(c.pre)
	}

	if e.HasError() {
		return result.Complete()
	}

	if c.child != nil {
		e.Evaluate(c.child)
	}

	if c.post != nil {
		e.Evaluate(c.post)
	}

	return result.Complete()
}
