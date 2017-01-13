package test

import (
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/ovrclk/gestalt"
	g "github.com/ovrclk/gestalt/builder"
)

func TestBG(t *testing.T) {
	//t.SkipNow()
	gestalt.Run(
		g.Suite("walker-dev").
			Run(g.SH("cleanup", "echo", "cleaning")).
			Run(
				g.Group("server").
					Run(g.BG().Run(g.SH("start", "while true; do echo .; sleep 1; done"))).
					Run(g.SH("ping", "sleep", "1"))).
			Run(g.SH("okay", "sleep 1")))
}

func TestParse(t *testing.T) {
	gestalt.Run(
		g.Suite("make-vars").
			Run(g.SH("producer", "echo", "foo", "bar", "baz").
				FN(g.P().Capture("a", "b", "c"))).
			Run(g.FN("consumer", readFields(t))))
}

func TestVars(t *testing.T) {

	producer := g.
		SH("producer", "echo", "foo", "bar", "baz").
		FN(g.P().Head().Capture("a", "b", "c"))

	consumer := g.
		FN("consumer", readFields(t))

	gestalt.Run(
		g.Suite("export-vars").
			Run(producer).
			Run(consumer))
}

func readFields(t *testing.T) func(gestalt.RunCtx) gestalt.Result {
	return func(rctx gestalt.RunCtx) gestalt.Result {
		values := rctx.Values()

		if len := len(rctx.Values()); len != 3 {
			t.Fatalf("incorrect values size (%v != %v)", len, 3)
		}

		if x := values["a"]; x != "foo" {
			t.Fatalf("incorrect values size (%v != %v)", x, "foo")
		}

		if x := values["b"]; x != "bar" {
			t.Fatalf("incorrect values size (%v != %v)", x, "bar")
		}

		if x := values["c"]; x != "baz" {
			t.Fatalf("incorrect values size (%v != %v)", x, "baz")
		}

		return gestalt.ResultSuccess()
	}
}

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	os.Exit(m.Run())
}
