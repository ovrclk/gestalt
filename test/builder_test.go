package test

import (
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/ovrclk/gestalt"
)

func TestBG(t *testing.T) {
	g := gestalt.RootBuilder()
	gestalt.Run(
		g.Suite("walker-dev").
			Run(g.SH("cleanup", "echo", "cleaning")).
			Run(
				g.Group("server").
					Run(g.SH("start", "while true; do echo .; sleep 1; done").BG()).
					Run(g.SH("ping", "sleep", "1"))).
			Run(g.SH("okay", "sleep 1")).
			Build())
}

func TestParse(t *testing.T) {

}

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	os.Exit(m.Run())
}
