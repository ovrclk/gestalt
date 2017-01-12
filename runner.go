package gestalt

import (
	"context"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
)

const (
	RunStateStopped RunState = iota
	RunStateRunning
	RunStateComplete
	RunStateError
)

type BuildCtx interface {
	Logger() logrus.FieldLogger
}

type RunCtx interface {
	Context() context.Context
	Logger() logrus.FieldLogger
}

type Runner interface {
	RunCtx
}

type Runable struct {
	Retries int
	Timeout time.Duration
	Exec    func(RunCtx) Result
}

func NullRunable() *Runable {
	return &Runable{
		Exec: func(bctx RunCtx) Result {
			return ResultSuccess()
		},
	}
}

type RunState int

type Result interface {
	State() RunState
	Err() error
	Values() ResultValues
}

type ResultValues map[string]interface{}

type result struct {
	state  RunState
	values ResultValues
	err    error
}

func NewResult(state RunState, values ResultValues, err error) *result {
	return &result{state, values, err}
}

func ResultSuccess() *result {
	return NewResult(RunStateComplete, nil, nil)
}

func ResultError(err error) *result {
	return NewResult(RunStateError, nil, err)
}

func (r *result) State() RunState {
	return r.state
}

func (r *result) Err() error {
	return r.err
}

func (r *result) Values() ResultValues {
	return r.values
}

type runner struct {
	name   string
	ctx    context.Context
	logger logrus.FieldLogger
}

func NewRunner() *runner {
	return &runner{"", context.TODO(), logrus.StandardLogger()}
}

func Run(c Component) error {
	return NewRunner().Run(c)
}

func (r *runner) Context() context.Context {
	return r.ctx
}

func (r *runner) Logger() logrus.FieldLogger {
	return r.logger
}

func (r *runner) cloneFor(c Component) *runner {
	name := fmt.Sprintf("%v/%v", r.name, c.Name())
	return &runner{
		ctx:    r.ctx,
		logger: r.logger.WithField("name", name),
		name:   name,
	}
}

func (r *runner) Run(c Component) error {
	r.logger.Infof("Running %v...", c.Name())
	err := r.cloneFor(c).startComponent(c)
	if err != nil {
		return err
	}
	return nil
}

func (r *runner) runComponent(c Component) error {
	runable := c.Build(r)
	result := runable.Exec(r)
	switch result.State() {
	case RunStateComplete:
		r.Logger().Debugf("Component %v complete: %v", c.Name(), result.Values())
	case RunStateError:
		r.Logger().WithError(result.Err()).Errorf("Error running %v", c.Name())
		return result.Err()
	case RunStateRunning:

	default:
		err := fmt.Errorf("Unknown state: %v", result.State())
		r.Logger().WithError(err).Errorf("Error running %v", c.Name())
		return err
	}
	for _, child := range c.Children() {
		if err := r.Run(child); err != nil {
			return err
		}
	}
	return nil
}

// create farm:
//  - reset-env
//  - start-walker
//   - wait-available
//  - create-farm
