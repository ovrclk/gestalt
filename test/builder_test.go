package test

import (
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/ovrclk/gestalt"
	g "github.com/ovrclk/gestalt/builder"
	"github.com/ovrclk/gestalt/result"
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
		FN(g.P().Head().Capture("a", "b", "c")).
		WithMeta(g.M().
			Export("a", "b", "c"))

		//Requires("foo", "bar")
		//Exports("a", "b", "c")

	consumer := g.
		FN("consumer", readFields(t)).
		WithMeta(g.M().
			Require("a", "b", "c"))

	suite := g.Suite("export-vars").
		Run(producer).
		Run(consumer)
		//ExportsFrom("producer").
		//RequiresFor("producer")

	gestalt.Run(suite)
}

func TestEnsure(t *testing.T) {
	ran := false
	result := gestalt.Run(g.Ensure("a").
		First(
			g.SH("producer", "echo", "foo", "bar", "baz").
				FN(g.P().Head().Capture("a", "b", "c"))).
		Run(
			g.SH("failing", "false")).
		Finally(
			g.FN("consumer", func(_ gestalt.Evaluator) result.Result {
				ran = true
				return result.Complete()
			})))
	if result == nil {
		t.Fatal("error result not returned")
	}
	if ran != true {
		t.Fatal("fnally block didn't run")
	}
}

func TestDump(t *testing.T) {
	//t.SkipNow()
	gestalt.Dump(g.
		Suite("a").
		Run(g.BG().Run(g.SH("b", "echo", "foo", "bar"))).
		Run(g.SH("c", "echo", "hello")).
		Run(g.Group("z").
			Run(g.Retry(10).Run(g.SH("x", "echo", "sup")))))
}

func readFields(t *testing.T) gestalt.Runable {
	return func(e gestalt.Evaluator) result.Result {

		vars := e.Vars()

		if count := vars.Count(); count != 3 {
			t.Fatalf("incorrect values size (%v != %v)", count, 3)
		}

		if x := vars.Get("a"); x != "foo" {
			t.Fatalf("incorrect values size (%v != %v)", x, "foo")
		}

		if x := vars.Get("b"); x != "bar" {
			t.Fatalf("incorrect values size (%v != %v)", x, "bar")
		}

		if x := vars.Get("c"); x != "baz" {
			t.Fatalf("incorrect values size (%v != %v)", x, "baz")
		}

		return result.Complete()
	}
}

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	os.Exit(m.Run())
}
