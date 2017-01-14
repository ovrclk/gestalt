package gestalt

import (
	"fmt"
	"time"

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
	Run(Component) CompositeComponent
}

type WrapComponent interface {
	Component
	Child() Component
	Run(Component) Component
}

type C struct {
	name     string
	terminal bool
	build    func(Builder) Runable
	meta     vars.Meta
}

type CC struct {
	C
	children []Component
}

type WC struct {
	C
	child Component
}

func (c *C) Name() string {
	return c.name
}

func (c *C) IsPassThrough() bool {
	return false
}

func (c *C) WithMeta(m vars.Meta) Component {
	return c
}

func (c *C) Eval(e Evaluator) result.Result {
	if c.build == nil {
		return result.Complete()
	}
	return c.build(e.Builder())(e)
}

func (c *C) Build(b Builder) Runable {
	if c.build == nil {
		return nil
	}
	return c.build(b)
}

func (c *CC) Children() []Component {
	return c.children
}

func (c *CC) WithMeta(m vars.Meta) Component {
	c.C.WithMeta(m)
	return c
}

func (c *CC) Run(child Component) CompositeComponent {
	c.children = append(c.children, child)
	return c
}

func (c *CC) Eval(e Evaluator) result.Result {

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

func (c *WC) IsPassThrough() bool {
	return true
}

func (c *WC) WithMeta(m vars.Meta) Component {
	c.C.WithMeta(m)
	return c
}

func (c *WC) Child() Component {
	return c.child
}

func (c *WC) Run(child Component) Component {
	c.child = child
	return c
}

func NewComponent(name string, fn func(Builder) Runable) *C {
	return &C{name: name, build: fn}
}

func NewComponentR(name string, fn Runable) *C {
	return NewComponent(name, func(bctx Builder) Runable {
		return fn
	})
}

func NewSuite(name string) *CC {
	return &CC{
		C: C{name: name, terminal: true},
	}
}

func NewGroup(name string) *CC {
	return &CC{
		C: C{name: name, terminal: false},
	}
}

func NewWrapComponent(name string, fn func(WrapComponent, Evaluator) result.Result) *WC {
	c := &WC{C: C{name: name}}
	c.build = func(bctx Builder) Runable {
		return func(e Evaluator) result.Result {
			return fn(c, e)
		}
	}
	return c
}

func NewRetryComponent(tries int, delay time.Duration) *WC {
	return NewWrapComponent(
		"retry",
		func(c WrapComponent, e Evaluator) result.Result {
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
	return NewWrapComponent("background", func(c WrapComponent, e Evaluator) result.Result {
		ch := make(chan result.Result)
		go func() {
			defer close(ch)
			ch <- e.Evaluate(c.Child()).Wait()
		}()
		return result.Running(func() result.Result {
			return <-ch
		})
	})
}
