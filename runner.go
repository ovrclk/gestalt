package gestalt

import (
	"context"
	"fmt"
	"sync"
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
	Wait()
}

type ResultValues map[string]interface{}

type result struct {
	state  RunState
	values ResultValues
	err    error
	wait   func()
}

func NewResult(state RunState, values ResultValues, err error) *result {
	return &result{state, values, err, nil}
}

func ResultSuccess() *result {
	return NewResult(RunStateComplete, nil, nil)
}

func ResultError(err error) *result {
	return NewResult(RunStateError, nil, err)
}

func ResultRunning(f func()) *result {
	return &result{RunStateRunning, nil, nil, f}
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

func (r *result) Wait() {
	if r.state != RunStateRunning && r.wait != nil {
		r.wait()
	}
}

type runner struct {
	name     string
	ctx      context.Context
	cancel   context.CancelFunc
	logger   logrus.FieldLogger
	wg       sync.WaitGroup
	children []*runner
}

func NewRunner() *runner {
	ctx, cancel := context.WithCancel(context.TODO())
	return &runner{
		name:   "",
		ctx:    ctx,
		cancel: cancel,
		logger: logrus.StandardLogger(),
	}
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
	ctx, cancel := context.WithCancel(r.ctx)
	child := &runner{
		ctx:    ctx,
		cancel: cancel,
		logger: r.logger.WithField("name", name),
		name:   name,
	}
	r.children = append(r.children, child)
	return child
}

func (r *runner) Wait() {
	for _, c := range r.children {
		c.Wait()
	}
	r.wg.Wait()
}

func (r *runner) Run(c Component) error {
	err := r.cloneFor(c).runComponent(c)
	if c.IsTerminal() || err != nil {
		r.cancel()
		r.Wait()
		return err
	}
	return nil
}

func (r *runner) runComponent(c Component) error {
	r.Logger().Infof("start")
	runable := c.Build(r)
	result := runable.Exec(r)
	switch result.State() {
	case RunStateComplete:
		r.Logger().Debugf("complete")
	case RunStateError:
		r.Logger().WithError(result.Err()).Errorf("error")
		return result.Err()
	case RunStateRunning:
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			result.Wait()
			r.Logger().Infof("complete")
		}()
	default:
		err := fmt.Errorf("Unknown state: %v", result.State())
		r.Logger().WithError(err).Errorf("error")
		return err
	}
	for _, child := range c.Children() {
		if err := r.Run(child); err != nil {
			return err
		}
	}
	return nil
}
