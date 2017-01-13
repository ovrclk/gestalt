package test

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/ovrclk/gestalt"
	g "github.com/ovrclk/gestalt/builder"
)

func TestBG(t *testing.T) {
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
				FN(produceFields)).
			Run(g.FN("consumer", readFields(t))))
}

func TestVars(t *testing.T) {

	producer := g.
		SH("producer", "echo", "foo", "bar", "baz").
		FN(produceFields)

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

func produceFields(b *bufio.Reader, rctx gestalt.RunCtx) (gestalt.ResultValues, error) {
	vals := make(gestalt.ResultValues)

	line, _, err := b.ReadLine()

	if err != nil {
		return vals, err
	}

	fields := strings.Fields(string(line))

	if len(fields) != 3 {
		return vals, fmt.Errorf("invalid format")
	}

	vals["a"] = fields[0]
	vals["b"] = fields[1]
	vals["c"] = fields[2]

	return vals, nil
}

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	os.Exit(m.Run())
}
