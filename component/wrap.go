package component

import (
	"fmt"
	"time"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/result"
	"github.com/ovrclk/gestalt/vars"
)

type Wrap interface {
	gestalt.Component
	Child() gestalt.Component
	Run(gestalt.Component) gestalt.Component
}

/* wrapped */
type WC struct {
	cmp     gestalt.Component
	wrapper WrapFn
	child   gestalt.Component
}

type WrapFn func(Wrap, gestalt.Evaluator) result.Result

func NewWrap(name string, wrapper WrapFn) *WC {
	return &WC{
		cmp:     gestalt.NewComponent(name, nil),
		wrapper: wrapper,
	}
}

func NewRetry(tries int, delay time.Duration) *WC {
	return NewWrap(
		"retry",
		func(c Wrap, e gestalt.Evaluator) result.Result {
			for i := 0; i < tries; i++ {
				if i > 0 {
					time.Sleep(delay)
				}
				res := e.Evaluate(c.Child())
				switch res.State() {
				case result.StateComplete, result.StateRunning:
					return res
				}
			}
			return result.Error(fmt.Errorf("too many retries"))
		})
}

func NewBG() *WC {
	return NewWrap("background", func(c Wrap, e gestalt.Evaluator) result.Result {
		return e.Fork(c.Child())
	})
}

func (c *WC) Eval(e gestalt.Evaluator) result.Result {
	return c.wrapper(c, e)
}

func (c *WC) IsPassThrough() bool {
	return true
}

func (c *WC) Name() string {
	return c.cmp.Name()
}

func (c *WC) Meta() vars.Meta {
	m := c.cmp.Meta()
	for _, child := range c.Children() {
		m = m.Merge(child.Meta())
	}
	return m
}

func (c *WC) WithMeta(m vars.Meta) gestalt.Component {
	c.cmp.WithMeta(m)
	return c
}

func (c *WC) Children() []gestalt.Component {
	return []gestalt.Component{c.Child()}
}

func (c *WC) Child() gestalt.Component {
	return c.child
}

func (c *WC) Run(child gestalt.Component) gestalt.Component {
	c.child = child
	return c
}
