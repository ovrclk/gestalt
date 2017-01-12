package gestalt

import (
	"context"
	"fmt"
	"sync"

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
	Values() ResultValues
}

type RunCtx interface {
	Context() context.Context
	Logger() logrus.FieldLogger
	Values() ResultValues
	Run(c Component) error
}

type Runner interface {
	RunCtx
}

type Runable func(RunCtx) Result

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
	if r.state == RunStateRunning && r.wait != nil {
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

	vals ResultValues
}

func NewRunner() *runner {
	ctx, cancel := context.WithCancel(context.TODO())
	return &runner{
		name:   "",
		ctx:    ctx,
		cancel: cancel,
		logger: logrus.StandardLogger(),
		vals:   make(ResultValues),
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

func (r *runner) Values() ResultValues {
	return r.vals
}

func (r *runner) cloneFor(c Component) *runner {

	name := r.name
	if cname := c.Name(); cname != "" {
		name = fmt.Sprintf("%v/%v", name, cname)
	}

	ctx, cancel := context.WithCancel(r.ctx)

	child := &runner{
		ctx:    ctx,
		cancel: cancel,
		logger: r.logger.WithField("name", name),
		name:   name,
	}

	r.children = append(r.children, child)

	if c.IsTerminal() {
		child.vals = make(ResultValues)
		for k, v := range r.vals {
			child.vals[k] = v
		}
	} else {
		child.vals = r.vals
	}
	return child
}

func (r *runner) Wait() {
	for _, c := range r.children {
		c.Wait()
	}
	r.wg.Wait()
}

func (r *runner) Run(c Component) error {
	return r.cloneFor(c).doRun(c)
}

func (r *runner) doRun(c Component) error {
	r.Logger().Infof("begin")

	err := r.runAll(c)

	if err != nil {
		r.stop()
		r.Logger().WithError(err).Errorf("end")
		return err
	}

	if c.IsTerminal() {
		r.stop()
	}

	r.Logger().Infof("end")
	return nil
}

func (r *runner) stop() {
	r.cancel()
	r.Wait()
}

func (r *runner) runAll(c Component) error {
	if err := r.buildAndRun(c); err != nil {
		return err
	}
	for _, child := range c.Children() {
		if err := r.Run(child); err != nil {
			return err
		}
	}
	return nil
}

func (r *runner) buildAndRun(c Component) error {
	fn := c.Build(r)

	if fn == nil {
		return nil
	}

	r.Logger().Infof("start")

	result := fn(r)
	switch result.State() {
	case RunStateComplete:
		r.Logger().Infof("complete")
		r.addVars(result.Values())
	case RunStateError:
		r.Logger().WithError(result.Err()).Errorf("error")
		return result.Err()
	case RunStateRunning:
		r.wg.Add(1)
		r.Logger().Infof("background")
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
	return nil
}

func (r *runner) addVars(vals ResultValues) {
	if vals != nil {
		r.Logger().Debugf("results: %v", vals)
		for k, v := range vals {
			r.vals[k] = v
		}
	}
}
