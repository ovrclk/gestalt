package component

import (
	"fmt"
	"time"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/result"
	"github.com/ovrclk/gestalt/vars"
)

type WrapComponent interface {
	gestalt.Component
	Child() gestalt.Component
	Run(gestalt.Component) gestalt.Component
}

/* wrapped */
type WC struct {
	gestalt.C
	wrapper WrapFn
	child   gestalt.Component
}

type WrapFn func(WrapComponent, gestalt.Evaluator) result.Result

func NewWrapComponent(name string, wrapper WrapFn) *WC {
	return &WC{
		C:       *gestalt.NewComponent(name, nil),
		wrapper: wrapper,
	}
}

func NewRetryComponent(tries int, delay time.Duration) *WC {
	return NewWrapComponent(
		"retry",
		func(c WrapComponent, e gestalt.Evaluator) result.Result {
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

func NewBGComponent() *WC {
	return NewWrapComponent("background", func(c WrapComponent, e gestalt.Evaluator) result.Result {
		return e.Fork(c.Child())
	})
}

func (c *WC) Eval(e gestalt.Evaluator) result.Result {
	return c.Build(e.Builder)(e)
}

func (c *WC) Build(_ gestalt.Builder) gestalt.Runable {
	return func(e gestalt.Evaluator) result.Result {
		return c.wrapper(c, e)
	}
}

func (c *WC) IsPassThrough() bool {
	return true
}

func (c *WC) Children() []gestalt.Component {
	return []gestalt.Component{c.Child()}
}

func (c *WC) WithMeta(m vars.Meta) gestalt.Component {
	c.C.WithMeta(m)
	return c
}

func (c *WC) Child() gestalt.Component {
	return c.child
}

func (c *WC) Run(child gestalt.Component) gestalt.Component {
	c.child = child
	return c
}
