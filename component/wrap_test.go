package component_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/component"
	"github.com/ovrclk/gestalt/result"
)

func TestRetry(t *testing.T) {
	count := 0

	check := gestalt.NewComponent("check", func(e gestalt.Evaluator) result.Result {
		assert.Equal(t, e.Path(), "/check")
		if count == 2 {
			return result.Complete()
		}
		count++
		return result.Error(fmt.Errorf("invalid count"))
	})

	{
		count = 0
		cmp := component.NewRetry(4, time.Millisecond).Run(check)
		res := gestalt.NewEvaluator().Evaluate(cmp)

		assert.True(t, res.IsComplete())
		assert.Nil(t, res.Err())

		assert.Equal(t, count, 2)
	}

	{
		count = 0
		cmp := component.NewRetry(2, time.Millisecond).Run(check)
		res := gestalt.NewEvaluator().Evaluate(cmp)
		assert.True(t, res.IsError())
		assert.Error(t, res.Err())
		assert.Equal(t, count, 2)
	}
}

func TestBG(t *testing.T) {
	server := gestalt.NewComponent("server", func(e gestalt.Evaluator) result.Result {
		select {
		case <-e.Context().Done():
			return result.Complete()
		case <-time.After(time.Second / 2):
			assert.Fail(t, "context channel never closed")
			return result.Error(fmt.Errorf("context channel never closed"))
		}
	})

	cmp := component.NewBG().Run(server)
	e := gestalt.NewEvaluator()
	res := e.Evaluate(cmp)
	assert.True(t, res.IsComplete())
	e.Stop()
	e.Wait()
}
