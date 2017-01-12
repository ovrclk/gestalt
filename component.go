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
	Children() []Component
	AddChild(Component) Component
	Build(BuildCtx) Runable
	IsTerminal() bool
}

type component struct {
	name     string
	children []Component
}

func NewComponent(name string) *component {
	return &component{name: name}
}

func (c *component) Name() string {
	return c.name
}

func (c *component) Children() []Component {
	return c.children
}

func (c *component) IsTerminal() bool {
	return false
}

func (c *component) AddChild(child Component) Component {
	c.children = append(c.children, child)
	return child
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

type RetryComponent struct {
	component
	child Component
	tries int
	delay time.Duration
}

func (c *RetryComponent) Name() string {
	return ""
}

func (c *RetryComponent) Build(bctx BuildCtx) Runable {
	return func(rctx RunCtx) Result {
		for i := 0; i < c.tries; i++ {
			if err := rctx.Run(c.child); err == nil {
				return ResultSuccess()
			}
			time.Sleep(c.delay)
		}
		return ResultError(fmt.Errorf("too many retries"))
	}
}

type BGComponent struct {
	component
	child Component
}

func (c *BGComponent) Name() string {
	return ""
}

func (c *BGComponent) Build(bctx BuildCtx) Runable {
	return func(rctx RunCtx) Result {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := rctx.Run(c.child); err != nil {
				rctx.Logger().WithError(err).Errorf("BG")
			}
		}()
		return ResultRunning(wg.Wait)
	}
}
