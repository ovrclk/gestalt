package builder

import (
	"time"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/component"
	"github.com/ovrclk/gestalt/exec"
	"github.com/ovrclk/gestalt/vars"
)

func Group(name string) component.CompositeComponent {
	return component.NewGroup(name)
}

func Suite(name string) component.CompositeComponent {
	return component.NewSuite(name)
}

func BG() component.WrapComponent {
	return component.NewBGComponent()
}

func Retry(tries int) component.WrapComponent {
	return component.NewRetryComponent(tries, time.Second)
}

func Ensure(name string) component.EnsureComponent {
	return component.NewEnsureComponent(name)
}

func FN(name string, action gestalt.Runable) gestalt.Component {
	return gestalt.NewComponent(name, action)
}

func SH(name, cmd string, args ...string) *exec.Cmd {
	return exec.SH(name, cmd, args...)
}

func EXEC(name, cmd string, args ...string) *exec.Cmd {
	return exec.EXEC(name, cmd, args...)
}

func P() exec.TextPipe {
	return exec.P()
}

func M() vars.Meta {
	return vars.NewMeta()
}
