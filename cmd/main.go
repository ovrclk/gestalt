package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/ovrclk/gestalt"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	s := gestalt.NewSuite("foo")
	s.AddChild(gestalt.NewShellComponent("ls", "ls", []string{"/"}))
	if err := gestalt.Run(s); err != nil {
		logrus.Errorf("Error running: %v", err)
	}
}
