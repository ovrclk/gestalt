package builder

import (
	"time"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/component"
	"github.com/ovrclk/gestalt/exec"
	"github.com/ovrclk/gestalt/vars"
)

func Group(name string) component.Group {
	return component.NewGroup(name)
}

func Suite(name string) component.Group {
	return component.NewSuite(name)
}

func Noop(name string) gestalt.Component {
	return gestalt.NoopComponent(name)
}

func BG() component.Wrap {
	return component.NewBG()
}

func Retry(tries int) component.Wrap {
	return component.NewRetry(tries, time.Second)
}

func Ensure(name string) component.Ensure {
	return component.NewEnsure(name)
}

func Ignore() component.Wrap {
	return component.NewIgnore()
}

func FN(name string, action gestalt.Action) gestalt.Component {
	return gestalt.NewComponent(name, action)
}

func SH(name, cmd string, args ...string) exec.Cmd {
	return exec.SH(name, cmd, args...)
}

func EXEC(name, cmd string, args ...string) exec.Cmd {
	return exec.EXEC(name, cmd, args...)
}

func Capture(columns ...string) exec.CmdFn {
	return exec.Capture(columns...)
}

func Columns(columns ...string) exec.Pipeline {
	return exec.ParseColumns(columns...)
}

func Require(args ...string) vars.Meta {
	return vars.NewMeta().Require(args...)
}

func Export(args ...string) vars.Meta {
	return vars.NewMeta().Export(args...)
}

func Default(k, v string) vars.Meta {
	return vars.NewMeta().Default(k, v)
}

func Ref(name string) vars.Ref {
	return vars.NewRef(name)
}

func Compose(steps ...component.Ensure) component.Ensure {
	return component.Compose(steps...)
}
