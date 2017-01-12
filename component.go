package gestalt

import (
	"os/exec"
	"strings"
)

type Component interface {
	Name() string
	Children() []Component
	AddChild(Component) Component
	Build(BuildCtx) *Runable
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

func (c *ShellComponent) Build(bctx BuildCtx) *Runable {
	return &Runable{
		Exec: func(rctx RunCtx) Result {
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

			go func() {
				buf := make([]byte, 80)
				for {
					n, err := stdout.Read(buf)
					if err != nil {
						break
					}
					rctx.Logger().Debug(string(buf[0:n]))
				}
			}()

			go func() {
				buf := make([]byte, 80)
				for {
					n, err := stderr.Read(buf)
					if err != nil {
						break
					}
					rctx.Logger().Error(string(buf[0:n]))
				}
			}()

			if c.bg {
				return ResultRunning(func() {
					if err := cmd.Wait(); err != nil {
						rctx.Logger().WithError(err).Error("command failed")
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
		},
	}
}
