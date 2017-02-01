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

func TestExportWithParentGroup(t *testing.T) {

	generator := exportComponent("a", "foo")
	check, ran := checkComponent(t, "a", "foo")

	cgroup := component.NewGroup("create")
	cgroup.Run(component.NewRetry(1, 0).Run(generator))
	cgroup.WithMeta(vars.NewMeta().Export("a"))

	pgroup := component.NewGroup("parent")
	pgroup.Run(cgroup)
	pgroup.Run(check)

	e := gestalt.NewEvaluator()
	e.Evaluate(pgroup)
	assert.True(t, *ran)
}

func TestExportWithEnsureParent(t *testing.T) {
	generator := exportComponent("a", "foo")
	check, ran := checkComponent(t, "a", "foo")
	fcheck, fran := checkComponent(t, "a", "foo")

	cgroup := component.NewGroup("create")
	cgroup.Run(component.NewRetry(1, 0).Run(generator))
	cgroup.WithMeta(vars.NewMeta().Export("a"))

	pgroup := component.NewEnsure("parent")
	pgroup.First(generator)
	pgroup.Run(check)
	pgroup.Finally(fcheck)

	e := gestalt.NewEvaluator()
	e.Evaluate(pgroup)

	assert.True(t, *ran)
	assert.True(t, *fran)
}

func exportComponent(key, value string) gestalt.Component {
	return gestalt.NewComponent("create", func(e gestalt.Evaluator) result.Result {
		e.Emit(key, value)
		return result.Complete()
	}).WithMeta(vars.NewMeta().Export(key))
}

func checkComponent(t *testing.T, key, value string) (gestalt.Component, *bool) {
	ran := false
	check := gestalt.NewComponent("check", func(e gestalt.Evaluator) result.Result {
		ran = true
		assert.True(t, e.Vars().Has(key))
		assert.Equal(t, value, e.Vars().Get(key))
		return result.Complete()
	}).WithMeta(vars.NewMeta().Require(key))
	return check, &ran
}
