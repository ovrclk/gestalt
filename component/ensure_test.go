package component_test

import (
	"fmt"
	"testing"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/component"
	"github.com/stretchr/testify/assert"
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
