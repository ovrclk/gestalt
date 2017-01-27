package component_test

import (
	"testing"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/component"
	"github.com/ovrclk/gestalt/result"
	"github.com/stretchr/testify/assert"
)

func TestGroup(t *testing.T) {
	ch := make(chan interface{})

	cmp := component.NewGroup("server")
	cmp.Run(component.NewBG().
		Run(gestalt.NewComponent("server", func(e gestalt.Evaluator) result.Result {
			select {
			case <-e.Context().Done():
				t.Error("context cancelled")
				return result.Complete()
			case <-ch:
				return result.Complete()
			}
		})))

	cmp.Run(gestalt.NewComponent("check", func(e gestalt.Evaluator) result.Result {
		return result.Complete()
	}))

	e := gestalt.NewEvaluator()
	res := e.Evaluate(cmp)

	assert.True(t, res.IsComplete())

	close(ch)

	e.Wait()
}

func TestSuite(t *testing.T) {
	ch := make(chan interface{})

	cmp := component.NewSuite("server")
	cmp.Run(component.NewBG().
		Run(gestalt.NewComponent("server", func(e gestalt.Evaluator) result.Result {
			select {
			case <-e.Context().Done():
				return result.Complete()
			case <-ch:
				t.Error("ch cancelled")
				return result.Complete()
			}
		})))

	cmp.Run(gestalt.NewComponent("check", func(e gestalt.Evaluator) result.Result {
		return result.Complete()
	}))

	e := gestalt.NewEvaluator()
	res := e.Evaluate(cmp)

	assert.True(t, res.IsComplete())

	close(ch)

	e.Wait()
}
