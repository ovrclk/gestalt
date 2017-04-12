package component_test

import (
	"fmt"
	"testing"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/component"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsure(t *testing.T) {
	cmp := component.NewEnsure("session")

	cmp.First(gestalt.NewComponent("signin", func(e gestalt.Evaluator) error {
		assert.Equal(t, "/signin", e.Path())
		return fmt.Errorf("can't create app")
	}))

	cmp.Run(gestalt.NewComponent("create-app", func(e gestalt.Evaluator) error {
		assert.Equal(t, "/create-app", e.Path())
		return fmt.Errorf("can't create app")
	}))

	ran := false
	cmp.First(gestalt.NewComponent("signout", func(e gestalt.Evaluator) error {
		assert.Equal(t, "/signout", e.Path())
		ran = true
		return nil
	}))

	e := gestalt.NewEvaluator()

	res := e.Evaluate(cmp)

	assert.True(t, ran)
	assert.NoError(t, res)
	assert.True(t, e.HasError())
}

func TestCompose(t *testing.T) {

	cmp1 := component.NewEnsure("session")
	cmp2 := component.NewEnsure("post-update")

	trace := make([]string, 0)

	cmp1.First(gestalt.NewComponent("signin", func(e gestalt.Evaluator) error {
		assert.Equal(t, "/signin", e.Path())
		trace = append(trace, "signin")
		return nil
	}))

	cmp1.Finally(gestalt.NewComponent("signout", func(e gestalt.Evaluator) error {
		assert.Equal(t, "/signout", e.Path())
		trace = append(trace, "signout")
		return nil
	}))

	cmp2.First(gestalt.NewComponent("post-update", func(e gestalt.Evaluator) error {
		assert.Equal(t, "/post-update", e.Path())
		trace = append(trace, "post-update")
		return nil
	}))

	cmp2.Finally(gestalt.NewComponent("delete-update", func(e gestalt.Evaluator) error {
		assert.Equal(t, "/delete-update", e.Path())
		trace = append(trace, "delete-update")
		return nil
	}))

	cmp := component.Compose(cmp1, cmp2)

	e := gestalt.NewEvaluator()

	require.NoError(t, e.Evaluate(cmp))
	require.Empty(t, e.Errors())

	require.Equal(t, []string{"signin", "post-update", "delete-update", "signout"}, trace)
}
