package test

import (
	"reflect"
	"testing"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/component"
	"github.com/ovrclk/gestalt/result"
	"github.com/ovrclk/gestalt/vars"
	"github.com/stretchr/testify/assert"
)

func TestMeta(t *testing.T) {
	exports := []string{"a"}

	c := gestalt.NewComponent("foo", nil).
		WithMeta(vars.NewMeta().Export(exports...))

	if x := c.Meta().Exports(); !reflect.DeepEqual(x, exports) {
		t.Errorf("meta export missing keys %v != %v", x, exports)
	}
}

func TestExport(t *testing.T) {

	generator := gestalt.NewComponent("create", func(e gestalt.Evaluator) result.Result {
		e.Emit("a", "foo")
		return result.Complete()
	}).WithMeta(vars.NewMeta().Export("a"))

	checkran := false
	check := gestalt.NewComponent("check", func(e gestalt.Evaluator) result.Result {
		checkran = true
		assert.True(t, e.Vars().Has("a"))
		assert.Equal(t, "foo", e.Vars().Get("a"))
		return result.Complete()
	}).WithMeta(vars.NewMeta().Require("a"))

	cgroup := component.NewGroup("create")
	cgroup.Run(component.NewRetry(1, 0).Run(generator))
	cgroup.WithMeta(vars.NewMeta().Export("a"))

	pgroup := component.NewGroup("parent")
	pgroup.Run(cgroup)
	pgroup.Run(check)

	e := gestalt.NewEvaluator()

	e.Evaluate(pgroup)

	assert.True(t, checkran)
}
