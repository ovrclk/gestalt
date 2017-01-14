package builder

import (
	"time"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/exec"
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

func FN(name string, fn gestalt.Runable) gestalt.Component {
	return gestalt.NewComponentR(name, fn)
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
