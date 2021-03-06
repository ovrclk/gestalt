package component

import (
	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/vars"
)

type Ensure interface {
	gestalt.Component
	First(gestalt.Component) Ensure
	Run(gestalt.Component) Ensure
	Finally(gestalt.Component) Ensure
}

type ensure struct {
	cmp gestalt.Component

	pre      gestalt.Component
	children []gestalt.Component
	post     gestalt.Component
}

func NewEnsure(name string) *ensure {
	return &ensure{
		cmp: gestalt.NewComponent(name, nil),
	}
}

func (c *ensure) First(child gestalt.Component) Ensure {
	c.pre = child
	return c
}

func (c *ensure) Run(child gestalt.Component) Ensure {
	c.children = append(c.children, child)
	return c
}

func (c *ensure) Finally(child gestalt.Component) Ensure {
	c.post = child
	return c
}

func (c *ensure) Name() string {
	return c.cmp.Name()
}

func (c *ensure) Meta() vars.Meta {
	return c.cmp.Meta()
}

func (c *ensure) WithMeta(m vars.Meta) gestalt.Component {
	c.cmp.WithMeta(m)
	return c
}

func (c *ensure) IsPassThrough() bool {
	return true
}

func (c *ensure) Children() []gestalt.Component {
	children := make([]gestalt.Component, 0)
	if c.pre != nil {
		children = append(children, c.pre)
	}
	children = append(children, c.children...)
	if c.post != nil {
		children = append(children, c.post)
	}
	return children
}

func (c *ensure) Eval(e gestalt.Evaluator) error {
	if c.pre != nil {
		e.Evaluate(c.pre)
	}

	if e.HasError() {
		return nil
	}

	for _, child := range c.children {
		e.Evaluate(child)
		if e.HasError() {
			break
		}
	}

	if c.post != nil {
		e.Evaluate(c.post)
	}

	return nil
}

// steps[0].Run(steps[1].Run(...steps[N]))
func Compose(steps ...Ensure) Ensure {
	count := len(steps)
	if count < 1 {
		return nil
	}
	current := steps[count-1]
	for i := count - 2; i >= 0; i-- {
		current = steps[i].Run(current)
	}
	return current
}
