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

func BG() component.Wrap {
	return component.NewBG()
}

func Retry(tries int) component.Wrap {
	return component.NewRetry(tries, time.Second)
}

func Ensure(name string) component.Ensure {
	return component.NewEnsure(name)
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

func Columns(columns ...string) exec.ObjectPipe {
	return exec.ParseColumns(columns...)
}

func P() exec.TextPipe {
	return exec.P()
}

func M() vars.Meta {
	return vars.NewMeta()
}
