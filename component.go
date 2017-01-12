package gestalt

import (
	"context"
	"io"
	"os/exec"
	"strings"
	"syscall"
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

type ShellComponent struct {
	component
	Path string
	Args []string
	Env  []string

	bg bool
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

func (c *ShellComponent) WithBG(bg bool) *ShellComponent {
	c.bg = bg
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

		go logStream(stdout, rctx.Logger().Debug)
		go logStream(stderr, rctx.Logger().Error)

		if c.bg {
			return ResultRunning(func() {
				if err := cmd.Wait(); err != nil {
					if !expectedExecError(err, rctx.Context()) {
						rctx.Logger().WithError(err).Error("command failed")
					}
				}
			})
		} else {
			if err := cmd.Wait(); err != nil {
				rctx.Logger().WithError(err).Error("command failed")
				return ResultError(err)
			} else {
				return ResultSuccess()
			}
		}
	}
}

func logStream(reader io.ReadCloser, log func(fmt ...interface{})) {
	buf := make([]byte, 80)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			break
		}

		log(strings.TrimRight(string(buf[0:n]), "\n\r"))
	}
}

// Fail silently if killed due to context being cancelled.
func expectedExecError(err error, ctx context.Context) bool {
	if exiterr, ok := err.(*exec.ExitError); ok && ctx != nil {
		if wstatus, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			if syscall.Signal(wstatus&0x7f) == syscall.SIGKILL {
				return true
			}
		}
	}
	return false
}
