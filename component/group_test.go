package component_test

import (
	"testing"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/component"
	"github.com/stretchr/testify/assert"
)

func TestGroup(t *testing.T) {
	ch := make(chan interface{})

	cmp := component.NewGroup("server")
	cmp.Run(component.NewBG().
		Run(gestalt.NewComponent("server", func(e gestalt.Evaluator) error {
			select {
			case <-e.Context().Done():
				t.Error("context cancelled")
				return nil
			case <-ch:
				return nil
			}
		})))

	cmp.Run(gestalt.NewComponent("check", func(e gestalt.Evaluator) error {
		return nil
	}))

	e := gestalt.NewEvaluator()
	res := e.Evaluate(cmp)

	assert.NoError(t, res)

	close(ch)

	e.Wait()
}

func TestSuite(t *testing.T) {
	ch := make(chan interface{})

	cmp := component.NewSuite("server")
	cmp.Run(component.NewBG().
		Run(gestalt.NewComponent("server", func(e gestalt.Evaluator) error {
			select {
			case <-e.Context().Done():
				return nil
			case <-ch:
				t.Error("ch cancelled")
				return nil
			}
		})))

	cmp.Run(gestalt.NewComponent("check", func(e gestalt.Evaluator) error {
		return nil
	}))

	e := gestalt.NewEvaluator()
	res := e.Evaluate(cmp)

	assert.NoError(t, res)

	close(ch)

	e.Wait()
}
