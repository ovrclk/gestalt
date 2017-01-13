package builder

import (
	"time"

	"github.com/ovrclk/gestalt"
)

func Group(name string) gestalt.CompositeComponent {
	return gestalt.NewGroup(name)
}

func Suite(name string) gestalt.CompositeComponent {
	return gestalt.NewSuite(name)
}

func BG() gestalt.WrapComponent {
	return gestalt.NewBGComponent()
}

func Retry(tries int) gestalt.WrapComponent {
	return gestalt.NewRetryComponent(tries, time.Second)
}

func FN(name string, fn func(rctx gestalt.RunCtx) gestalt.Result) gestalt.Component {
	return gestalt.NewComponentR(name, fn)
}

func SH(name, cmd string, args ...string) *gestalt.ShellComponent {
	return gestalt.SH(name, cmd, args...)
}

func EXEC(name, cmd string, args ...string) *gestalt.ShellComponent {
	return gestalt.EXEC(name, cmd, args...)
}
