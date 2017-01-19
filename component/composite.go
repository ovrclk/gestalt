package component

import (
	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/result"
	"github.com/ovrclk/gestalt/vars"
)

type CompositeComponent interface {
	gestalt.CompositeComponent
	Run(gestalt.Component) CompositeComponent
}

/* composite */
type CC struct {
	cmp      *gestalt.C
	terminal bool
	children []gestalt.Component
}

func NewSuite(name string) *CC {
	return &CC{
		cmp:      gestalt.NewComponent(name, nil),
		terminal: true,
	}
}

func NewGroup(name string) *CC {
	return &CC{
		cmp: gestalt.NewComponent(name, nil),
	}
}

func (c *CC) Children() []gestalt.Component {
	return c.children
}

func (c *CC) Name() string {
	return c.cmp.Name()
}

func (c *CC) Meta() vars.Meta {
	return c.cmp.Meta()
}

func (c *CC) WithMeta(m vars.Meta) gestalt.Component {
	c.cmp.WithMeta(m)
	return c
}

func (c *CC) IsPassThrough() bool {
	return false
}

func (c *CC) Run(child gestalt.Component) CompositeComponent {
	c.children = append(c.children, child)
	return c
}

func (c *CC) Eval(e gestalt.Evaluator) result.Result {

	rset := result.NewSet()

	// evaluate children up to an error
	for _, child := range c.Children() {
		rset.Add(e.Evaluate(child))
		if rset.IsError() {
			break
		}
	}

	if rset.IsError() || c.terminal {
		e.Stop()
		return rset.Wait()
	}

	return rset.Result()
}
