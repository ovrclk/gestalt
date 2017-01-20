package test

import (
	"testing"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/component"
	"github.com/ovrclk/gestalt/result"
	"github.com/ovrclk/gestalt/vars"
)

func noop(name string) gestalt.Component {
	return gestalt.NewComponent(name, func(_ gestalt.Evaluator) result.Result {
		return result.Complete()
	})
}

func TestValidate(t *testing.T) {
	suite := component.NewSuite("top")

	if missing := gestalt.Validate(suite); len(missing) > 0 {
		t.Errorf("missing vars for empty suite")
	}

	suite.Run(noop("a"))

	if missing := gestalt.Validate(suite); len(missing) > 0 {
		t.Fail()
	}

	suite.Run(noop("b").
		WithMeta(vars.NewMeta().Export("x")))

	if missing := gestalt.Validate(suite); len(missing) > 0 {
		t.Fail()
	}

	suite.Run(noop("c").
		WithMeta(vars.NewMeta().Require("x")))

	if missing := gestalt.Validate(suite); len(missing) > 0 {
		t.Fail()
	}

	suite.Run(noop("d").
		WithMeta(vars.NewMeta().Require("y")))

	if missing := gestalt.Validate(suite); len(missing) != 1 {
		t.Errorf("invalid missing vars count: %#v", missing)
	} else if missing[0].Path != "/top/d" || missing[0].Name != "y" {
		t.Errorf("invalid missing vars: %#v", missing)
	}

}
