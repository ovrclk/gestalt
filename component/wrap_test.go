package component_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/component"
)

func TestRetry(t *testing.T) {
	count := 0

	check := gestalt.NewComponent("check", func(e gestalt.Evaluator) error {
		assert.Equal(t, e.Path(), "/check")
		if count == 2 {
			return nil
		}
		count++
		return fmt.Errorf("invalid count")
	})

	{
		count = 0
		cmp := component.NewRetry(4, time.Millisecond).Run(check)
		res := gestalt.NewEvaluator().Evaluate(cmp)

		assert.NoError(t, res)
		assert.Equal(t, count, 2)
	}

	{
		count = 0
		cmp := component.NewRetry(2, time.Millisecond).Run(check)
		res := gestalt.NewEvaluator().Evaluate(cmp)
		assert.Error(t, res)
		assert.Equal(t, count, 2)
	}
}

func TestBG(t *testing.T) {
	server := gestalt.NewComponent("server", func(e gestalt.Evaluator) error {
		select {
		case <-e.Context().Done():
			return nil
		case <-time.After(time.Second / 2):
			assert.Fail(t, "context channel never closed")
			return fmt.Errorf("context channel never closed")
		}
	})

	cmp := component.NewBG().Run(server)
	e := gestalt.NewEvaluator()
	res := e.Evaluate(cmp)
	assert.NoError(t, res)
	e.Stop()
	e.Wait()
}

func TestIgnore(t *testing.T) {
	ran := false
	cmp := component.NewGroup("test").
		Run(component.NewIgnore().Run(errComponent("die"))).
		Run(gestalt.NewComponent("run", func(_ gestalt.Evaluator) error {
			ran = true
			return nil
		}))

	e := gestalt.NewEvaluator()
	res := e.Evaluate(cmp)

	assert.NoError(t, res)
	assert.True(t, ran)

}

func errComponent(msg string) gestalt.Component {
	return gestalt.NewComponent("failing", func(_ gestalt.Evaluator) error {
		return fmt.Errorf("failed: %v", msg)
	})
}
