package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/ovrclk/gestalt"
)

func WalkerDev() gestalt.Component {
	g := gestalt.NewGroup("walker-dev")
	g.AddChild(gestalt.SH("cleanup", "echo", "cleaning..."))
	g.AddChild(
		gestalt.SH("start", "while true; do echo .; sleep 1; done").WithBG(true)).
		AddChild(g.AddChild(gestalt.SH("ping", "sleep", "3")))
	return g
}

func CreateFarm(name string) gestalt.Component {
	g := gestalt.NewGroup("create-farm")
	g.AddChild(gestalt.SH("create", "echo", "walker farms:create", name))
	g.AddChild(gestalt.SH("check", "echo", "walker farms"))
	return g
}

func RemoveFarm(name string) gestalt.Component {
	g := gestalt.NewGroup("remove-farm")
	g.AddChild(gestalt.SH("remove", "echo", "walker farms:down", name))
	g.AddChild(gestalt.SH("check", "echo", "walker farms"))
	return g
}

func FarmSuite() gestalt.Component {
	name := "foo"
	s := gestalt.NewSuite("farms")
	s.AddChild(WalkerDev())
	s.AddChild(CreateFarm(name))
	s.AddChild(RemoveFarm(name))
	return s
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	s := FarmSuite()
	if err := gestalt.Run(s); err != nil {
		logrus.Errorf("Error running: %v", err)
	}
}
