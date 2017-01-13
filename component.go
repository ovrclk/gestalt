package gestalt

import (
	"fmt"
	"sync"
	"time"
)

type Component interface {
	Name() string
	IsTerminal() bool
	IsPassThrough() bool

	Build(BuildCtx) Runable

	Exports(...string) Component
	Imports(...string) Component
	Requires(...string) Component
}

type CompositeComponent interface {
	Component
	Children() []Component
	Run(Component) CompositeComponent

	ExportsFrom(...string) CompositeComponent
	ImportsFor(...string) CompositeComponent
	RequiresFor(...string) CompositeComponent
}

type WrapComponent interface {
	Component
	Child() Component
	Run(Component) Component
}

type C struct {
	name     string
	terminal bool
	build    func(BuildCtx) Runable
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

func (c *C) IsTerminal() bool {
	return c.terminal
}

func (c *C) IsPassThrough() bool {
	return false
}

func (c *C) Exports(names ...string) Component {
	return c
}
func (c *C) Imports(names ...string) Component {
	return c
}
func (c *C) Requires(names ...string) Component {
	return c
}

func (c *C) Build(bctx BuildCtx) Runable {
	if c.build == nil {
		return nil
	}
	return c.build(bctx)
}

func (c *CC) Children() []Component {
	return c.children
}

func (c *CC) Run(child Component) CompositeComponent {
	c.children = append(c.children, child)
	return c
}

func (c *CC) Exports(names ...string) Component {
	c.C.Exports(names...)
	return c
}
func (c *CC) Imports(names ...string) Component {
	c.C.Imports(names...)
	return c
}
func (c *CC) Requires(names ...string) Component {
	c.C.Requires(names...)
	return c
}

func (c *CC) ExportsFrom(names ...string) CompositeComponent {
	return c
}

func (c *CC) ImportsFor(names ...string) CompositeComponent {
	return c
}

func (c *CC) RequiresFor(names ...string) CompositeComponent {
	return c
}

func (c *WC) IsPassThrough() bool {
	return true
}

func (c *WC) Child() Component {
	return c.child
}

func (c *WC) Exports(names ...string) Component {
	c.C.Exports(names...)
	return c
}
func (c *WC) Imports(names ...string) Component {
	c.C.Imports(names...)
	return c
}
func (c *WC) Requires(names ...string) Component {
	c.C.Requires(names...)
	return c
}

func (c *WC) Run(child Component) Component {
	c.child = child
	return c
}

func NewComponent(name string, fn func(BuildCtx) Runable) *C {
	return &C{name: name, build: fn}
}

func NewComponentR(name string, fn func(RunCtx) Result) *C {
	return NewComponent(name, func(bctx BuildCtx) Runable {
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

func NewWrapComponent(name string, fn func(WrapComponent, RunCtx) Result) *WC {
	c := &WC{C: C{name: name}}
	c.build = func(bctx BuildCtx) Runable {
		return func(rctx RunCtx) Result {
			return fn(c, rctx)
		}
	}
	return c
}

func NewRetryComponent(tries int, delay time.Duration) *WC {
	return NewWrapComponent(
		"retry",
		func(c WrapComponent, rctx RunCtx) Result {
			for i := 0; i < tries; i++ {
				if err := rctx.Run(c.Child()); err == nil {
					return ResultSuccess()
				}
				time.Sleep(delay)
			}
			return ResultError(fmt.Errorf("too many retries"))
		})
}

func NewBGComponent() *WC {
	return NewWrapComponent("background", func(c WrapComponent, rctx RunCtx) Result {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := rctx.Run(c.Child()); err != nil {
				rctx.Logger().WithError(err).Errorf("BG")
			}
		}()
		return ResultRunning(wg.Wait)
	})
}
