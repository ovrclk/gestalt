package gestalt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Component interface {
	Name() string
	IsTerminal() bool
	Build(BuildCtx) Runable
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

type component struct {
	name     string
	terminal bool
	build    func(BuildCtx) Runable
}

type compositeComponent struct {
	component
	children []Component
}

type wrapComponent struct {
	component
	child Component
}

func (c *component) Name() string {
	return c.name
}

func (c *component) IsTerminal() bool {
	return c.terminal
}

func (c *component) Build(bctx BuildCtx) Runable {
	if c.build == nil {
		return nil
	}
	return c.build(bctx)
}

func (c *compositeComponent) Children() []Component {
	return c.children
}

func (c *compositeComponent) Run(child Component) CompositeComponent {
	c.children = append(c.children, child)
	return c
}

func (c *wrapComponent) Child() Component {
	return c.child
}

func (c *wrapComponent) Run(child Component) Component {
	c.child = child
	return c
}

func NewSuite(name string) CompositeComponent {
	return &compositeComponent{
		component: component{name: name, terminal: true},
	}
}

func NewGroup(name string) CompositeComponent {
	return &compositeComponent{
		component: component{name: name, terminal: false},
	}
}

func NewWrapComponent(fn func(WrapComponent, RunCtx) Result) WrapComponent {
	c := &wrapComponent{}
	c.build = func(bctx BuildCtx) Runable {
		return func(rctx RunCtx) Result {
			return fn(c, rctx)
		}
	}
	return c
}

func NewComponent(name string, fn func(BuildCtx) Runable) Component {
	return &component{name: name, build: fn}
}

func NewComponentR(name string, fn func(RunCtx) Result) Component {
	return NewComponent(name, func(bctx BuildCtx) Runable {
		return fn
	})
}

func NewRetryComponent(tries int, delay time.Duration) WrapComponent {
	return NewWrapComponent(
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

func NewBGComponent() WrapComponent {
	return NewWrapComponent(func(c WrapComponent, rctx RunCtx) Result {
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

type ShellHandler func(*bufio.Reader, RunCtx) (ResultValues, error)

type ShellComponent struct {
	component
	Path string
	Args []string
	Env  []string

	fn ShellHandler
}

func NewShellComponent(name string, path string, args []string) *ShellComponent {
	return &ShellComponent{
		component: component{name: name},
		Path:      path,
		Args:      args,
	}
}

func EXEC(name, path string, args ...string) *ShellComponent {
	return NewShellComponent(name, path, args)
}

func SH(name, cmd string, args ...string) *ShellComponent {
	return NewShellComponent(
		name,
		"/bin/sh",
		[]string{
			"-c",
			strings.Join(append([]string{cmd}, args...), " "),
		})
}

func (c *ShellComponent) FN(fn ShellHandler) *ShellComponent {
	c.fn = fn
	return c
}

func (c *ShellComponent) Build(bctx BuildCtx) Runable {
	return func(rctx RunCtx) Result {
		cmd := exec.CommandContext(rctx.Context(), c.Path, c.Args...)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return ResultError(err)
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return ResultError(err)
		}

		rctx.Logger().Debugf("running %v %v", cmd.Path, cmd.Args)
		if err := cmd.Start(); err != nil {
			rctx.Logger().WithError(err).Errorf("error running %v", cmd.Path)
			return ResultError(err)
		}

		var buf *bytes.Buffer
		if c.copyStdout() {
			buf = new(bytes.Buffer)
		}

		go logStream(stdout, rctx.Logger().Debug, buf)
		go logStream(stderr, rctx.Logger().Error, nil)

		if err := cmd.Wait(); err != nil {
			if !expectedExecError(err, rctx) {
				rctx.Logger().WithError(err).Error("command failed")
				return ResultError(err)
			}
		}

		if c.copyStdout() {
			vals, err := c.fn(bufio.NewReader(buf), rctx)
			if err != nil {
				return NewResult(RunStateError, vals, err)
			}
			return NewResult(RunStateComplete, vals, nil)
		}

		return ResultSuccess()
	}
}

func (c *ShellComponent) copyStdout() bool {
	return c.fn != nil
}

func logStream(reader io.ReadCloser, log func(fmt ...interface{}), b *bytes.Buffer) {
	buf := make([]byte, 80)
	for {
		n, err := reader.Read(buf)

		if b != nil && n > 0 {
			b.Write(buf[0:n])
		}

		if err != nil {
			break
		}

		log(strings.TrimRight(string(buf[0:n]), "\n\r"))
	}
}

// Fail silently if killed due to context being cancelled.
func expectedExecError(err error, rctx RunCtx) bool {
	if exiterr, ok := err.(*exec.ExitError); ok && rctx.Context().Err() != nil {
		if wstatus, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			if wstatus.Signal() == syscall.SIGKILL {
				return true
			}
		}
	}
	return false
}
