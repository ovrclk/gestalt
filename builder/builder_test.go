package builder_test

import (
	"testing"

	"github.com/ovrclk/gestalt"
	g "github.com/ovrclk/gestalt/builder"
)

func TestBG(t *testing.T) {
	//t.SkipNow()
	runComponent(t,
		g.Suite("walker-dev").
			Run(g.SH("cleanup", "echo", "cleaning")).
			Run(
				g.Group("server").
					Run(g.BG().Run(g.SH("start", "while true; do echo .; sleep 1; done"))).
					Run(g.SH("ping", "sleep", "1"))).
			Run(g.SH("okay", "sleep 1")))
}

func TestVars_siblings(t *testing.T) {
	suite := g.Suite("siblings").
		Run(producer(t)).
		Run(consumer(t))
	runComponent(t, suite)
}

func TestVars_passthrough(t *testing.T) {
	suite := g.Suite("has-passthru").
		Run(
			g.Retry(1).
				Run(producer(t))).
		Run(consumer(t))

	runComponent(t, suite)
}

func TestVars_group(t *testing.T) {
	suite := g.Suite("has-embedded").
		Run(g.Group("has-passthru").
			Run(producer(t)).
			WithMeta(g.Export("a", "b", "c"))).
		Run(consumer(t))
	runComponent(t, suite)
}

func TestVars_suite(t *testing.T) {
	suite := g.Suite("has-embedded").
		Run(g.Suite("has-passthru").
			Run(producer(t)).
			WithMeta(g.Export("a", "b", "c"))).
		Run(consumer(t))
	runComponent(t, suite)
}

func TestEnsure(t *testing.T) {
	//t.SkipNow()
	ran := false
	c := g.Ensure("a").
		First(producer(t)).
		Run(g.SH("failing", "false")).
		Finally(
			g.FN("consumer", func(_ gestalt.Evaluator) error {
				ran = true
				return nil
			}))

	assertGestaltFails(t, c, []string{})

	if ran != true {
		t.Fatal("fnally block didn't run")
	}
}

func TestCliVars(t *testing.T) {
	args := []string{
		"-sa=foo",
		"-sb=bar",
		"-sc=baz",
	}

	assertGestaltSuccess(t, consumer(t), args)
}

func TextExpand(t *testing.T) {
	producer := g.SH("producer", "echo", "{{host}}", "bar", "baz").
		FN(g.Capture("a", "b", "c")).
		WithMeta(g.Export("a", "b", "c").Require("host"))

	suite := g.Suite("suite").
		Run(producer).
		Run(consumer(t)).
		WithMeta(g.Require("host"))

	{
		args := []string{"-shost=foo"}
		assertGestaltSuccess(t, suite, args)
	}

	{
		args := []string{"-shost=bar"}
		assertGestaltFails(t, suite, args)
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

func producer(t *testing.T) gestalt.Component {
	return g.SH("producer", "echo", "foo", "bar", "baz").
		FN(g.Capture("a", "b", "c")).
		WithMeta(g.Export("a", "b", "c"))
}

func consumer(t *testing.T) gestalt.Component {
	return g.FN("consumer", readFields(t)).
		WithMeta(g.Require("a", "b", "c"))
}

func readFields(t *testing.T) gestalt.Action {
	return func(e gestalt.Evaluator) error {

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

		return nil
	}
}

func runComponent(t *testing.T, c gestalt.Component) {
	assertGestaltSuccess(t, c, []string{})
}

func assertGestaltSuccess(t *testing.T, c gestalt.Component, args []string) {
	terminate := func(status int) {
		if status != 0 {
			t.Fatalf("gestalt exited with nonzero status (%v", status)
		}
	}

	args = append([]string{"eval"}, args...)

	runner := gestalt.NewRunner().
		WithComponent(c).
		WithArgs(args).
		WithTerminate(terminate)
	runner.Run()
}

func assertGestaltFails(t *testing.T, c gestalt.Component, args []string) {
	terminate := func(status int) {
		if status == 0 {
			t.Fail()
		}
	}
	args = append([]string{"eval"}, args...)

	runner := gestalt.NewRunner().
		WithComponent(c).
		WithArgs(args).
		WithTerminate(terminate)

	runner.Run()
}
