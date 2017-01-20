package component

import (
	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/result"
	"github.com/ovrclk/gestalt/vars"
)

type EnsureComponent interface {
	gestalt.Component
	First(gestalt.Component) EnsureComponent
	Run(gestalt.Component) EnsureComponent
	Finally(gestalt.Component) EnsureComponent
}

type EC struct {
	cmp gestalt.Component

	pre   gestalt.Component
	child gestalt.Component
	post  gestalt.Component
}

func NewEnsureComponent(name string) *EC {
	return &EC{
		cmp: gestalt.NewComponent(name, nil),
	}
}

func (c *EC) First(child gestalt.Component) EnsureComponent {
	c.pre = child
	return c
}

func (c *EC) Run(child gestalt.Component) EnsureComponent {
	c.child = child
	return c
}

func (c *EC) Finally(child gestalt.Component) EnsureComponent {
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
	rset := result.NewSet()

	if c.pre != nil {
		rset.Add(e.Evaluate(c.pre))
		if rset.IsError() {
			return rset.Result()
		}
	}

	if c.child != nil {
		rset.Add(e.Evaluate(c.child))
	}

	if c.post != nil {
		rset.Add(e.Evaluate(c.post))
	}

	return rset.Result()
}
