package test

import (
	"testing"

	"github.com/ovrclk/gestalt"
)

func buildWalker() gestalt.Component {
	g := gestalt.RootBuilder()
	return g.Group("walker-dev").
		Run(g.SH("cleanup", "echo", "cleaning")).
		Run(g.SH("start", "while true; do echo .; sleep 1; done").BG().
			Run(g.SH("ping", "sleep", "1")))
}

func withVars() gestalt.Component {
	g := gestalt.RootBuilder()
	return g.Group("walker-dev").
		Run(g.SH("cleanup", "echo", "cleaning")).
		Run(g.SH("start", "while true; do echo .; sleep 1; done").BG().
			Run(g.SH("ping", "sleep", "1"))).
		Exports("cleanup.foo").
		ExportsAs("start.bar", "baz").
		Requires("host").
		Connects("host","start.ping.host").
}

func TestWalker(t *testing.T) {
	gestalt.Run(ruildWalker())
}
