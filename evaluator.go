package gestalt

import (
	"context"
	"fmt"

	"github.com/Sirupsen/logrus"
)

type Builder interface {
}
type Runable func(Evaluator) Result

func Run(node Component) error {
	return NewEvaluator().Evaluate(node).Wait().Err()
}

type Evaluator interface {
	Log() logrus.FieldLogger
	Evaluate(Component) Result
	Context() context.Context
	Builder() Builder
	Stop()
}

type evaluator struct {
	path   string
	ctx    context.Context
	cancel context.CancelFunc
	log    logrus.FieldLogger
}

func NewEvaluator() *evaluator {
	ctx, cancel := context.WithCancel(context.TODO())
	return &evaluator{
		path:   "",
		ctx:    ctx,
		cancel: cancel,
		log:    logrus.StandardLogger(),
	}
}

func (e *evaluator) Builder() Builder {
	return nil
}

func (e *evaluator) Log() logrus.FieldLogger {
	return e.log
}

func (e *evaluator) Context() context.Context {
	return e.ctx
}

func (e *evaluator) Stop() {
	e.cancel()
}

func (e *evaluator) Evaluate(node Component) Result {
	child := e.cloneFor(node)

	child.Log().Debug("start")
	result := node.Eval(child)
	child.Log().Debugf("end -> %v", result)

	return result
}

func (e *evaluator) cloneFor(node Component) *evaluator {

	path := e.path
	if !node.IsPassThrough() {
		path = fmt.Sprintf("%v/%v", path, node.Name())
	}

	ctx, cancel := context.WithCancel(e.ctx)

	return &evaluator{
		path:   path,
		ctx:    ctx,
		cancel: cancel,
		log:    e.log.WithField("path", path),
	}
}
