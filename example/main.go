package main

import (
	"github.com/ovrclk/gestalt"
	g "github.com/ovrclk/gestalt/builder"
)

func main() {
	c := suite()
	gestalt.Run(c)
}

func suite() gestalt.Component {
	return g.Suite("integration").
		Run(pingServer())
	//Run(failing())
}

func pingServer() gestalt.Component {
	return g.Group("server").
		Run(g.Retry(5).
			Run(
				g.Group("start").
					Run(g.SH("cleanup", "echo", "ping")).
					Run(g.SH("start", "echo", "start")))).
		Run(g.SH("check", "echo", "ok"))
	//	WithMeta(g.Require("host"))
}

func failing() gestalt.Component {
	return g.SH("mkdir", "mkdir", "/foo")
}
