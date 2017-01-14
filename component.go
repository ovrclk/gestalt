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

	Eval(Evaluator) Result
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

func (c *C) Eval(e Evaluator) Result {
	if c.build == nil {
		return ResultSuccess()
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

func (c *CC) Run(child Component) CompositeComponent {
	c.children = append(c.children, child)
	return c
}

func (c *WC) IsPassThrough() bool {
	return true
}

func (c *CC) Eval(e Evaluator) Result {
	results := make([]Result, 0)
	var result Result

	for _, child := range c.Children() {
		result = e.Evaluate(child)
		results = append(results, result)
		if result.State() == RunStateError {
			break
		}
	}

	if result.State() == RunStateError || c.terminal {
		e.Stop()
		for i, _ := range results {
			results[i] = results[i].Wait()
		}
	}

	errors := make([]error, 0)
	running := false

	for _, result := range results {
		switch result.State() {
		case RunStateError:
			errors = append(errors, result.Err())
		case RunStateRunning:
			running = true
		}
	}

	if len(errors) > 0 {
		return ResultError(fmt.Errorf("error running %v children", len(errors)))
	} else if running {
		return ResultRunning(func() Result {
			errors := make([]error, 0)
			for _, result := range results {
				final := result.Wait()
				if final.State() == RunStateError {
					errors = append(errors, final.Err())
				}
			}
			if len(errors) > 0 {
				return ResultError(fmt.Errorf("error running %v children", len(errors)))
			} else {
				return ResultSuccess()
			}
		})
	} else {
		return ResultSuccess()
	}
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

func NewWrapComponent(name string, fn func(WrapComponent, Evaluator) Result) *WC {
	c := &WC{C: C{name: name}}
	c.build = func(bctx Builder) Runable {
		return func(e Evaluator) Result {
			return fn(c, e)
		}
	}
	return c
}

func NewRetryComponent(tries int, delay time.Duration) *WC {
	return NewWrapComponent(
		"retry",
		func(c WrapComponent, e Evaluator) Result {
			for i := 0; i < tries; i++ {
				if i > 0 {
					time.Sleep(delay)
				}
				result := e.Evaluate(c.Child())
				switch result.State() {
				case RunStateComplete, RunStateRunning:
					return result
				}
			}
			return ResultError(fmt.Errorf("too many retries"))
		})
}

func NewBGComponent() *WC {
	return NewWrapComponent("background", func(c WrapComponent, e Evaluator) Result {
		var wg sync.WaitGroup
		var result Result
		wg.Add(1)
		go func() {
			defer wg.Done()
			result = e.Evaluate(c.Child()).Wait()
		}()
		return ResultRunning(func() Result {
			wg.Wait()
			return result
		})
	})
}
