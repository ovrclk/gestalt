package gestalt

import (
	"context"
	"fmt"
	"sync"

	"github.com/Sirupsen/logrus"
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

type runner struct {
	name     string
	ctx      context.Context
	cancel   context.CancelFunc
	logger   logrus.FieldLogger
	wg       sync.WaitGroup
	children []*runner
	vals     ResultValues
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
	if !c.PassThrough() {
		name = fmt.Sprintf("%v/%v", name, c.Name())
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

	if c, ok := c.(CompositeComponent); ok {
		for _, child := range c.Children() {
			if err := r.Run(child); err != nil {
				return err
			}
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
