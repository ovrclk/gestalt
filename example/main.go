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
}

func pingServer() gestalt.Component {
	return g.SH("ping", "echo", "ping")
	//	WithMeta(vars.NewMeta().Require("host"))
}
