package gestalt

import "os/exec"

type Component interface {
	Name() string
	Children() []Component
	AddChild(Component) Component
	Build(BuildCtx) *Runable
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

func (c *component) AddChild(child Component) Component {
	c.children = append(c.children, child)
	return child
}

type ShellComponent struct {
	component
	Path string
	Args []string
	Env  []string
}

func NewShellComponent(name string, path string, args []string) *ShellComponent {
	return &ShellComponent{
		component: component{name: name},
		Path:      path,
		Args:      args,
	}
}

func (c *ShellComponent) Build(bctx BuildCtx) *Runable {
	return &Runable{
		Exec: func(rctx RunCtx) Result {
			cmd := exec.CommandContext(rctx.Context(), c.Path, c.Args...)
			rctx.Logger().Debugf("running %v %v", cmd.Path, cmd.Args)
			if err := cmd.Run(); err != nil {
				rctx.Logger().WithError(err).Errorf("error running %v", cmd.Path)
				return ResultError(err)
			}
			return ResultSuccess()
		},
	}
}
