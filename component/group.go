package component

import (
	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/vars"
)

type Group interface {
	gestalt.CompositeComponent
	Run(gestalt.Component) Group
}

/* group component */
type GC struct {
	cmp      gestalt.Component
	terminal bool
	children []gestalt.Component
}

func NewSuite(name string) *GC {
	return &GC{
		cmp:      gestalt.NewComponent(name, nil),
		terminal: true,
	}
}

func NewGroup(name string) *GC {
	return &GC{
		cmp: gestalt.NewComponent(name, nil),
	}
}

func (c *GC) Children() []gestalt.Component {
	return c.children
}

func (c *GC) Name() string {
	return c.cmp.Name()
}

func (c *GC) Meta() vars.Meta {
	return c.cmp.Meta()
}

func (c *GC) WithMeta(m vars.Meta) gestalt.Component {
	c.cmp.WithMeta(m)
	return c
}

func (c *GC) IsPassThrough() bool {
	return false
}

func (c *GC) Run(child gestalt.Component) Group {
	c.children = append(c.children, child)
	return c
}

func (c *GC) Eval(e gestalt.Evaluator) error {

	// evaluate children up to an error
	for _, child := range c.Children() {
		e.Evaluate(child)
		if e.HasError() {
			break
		}
	}

	if e.HasError() || c.terminal {
		e.Stop()
		e.Wait()
	}

	return nil
}
