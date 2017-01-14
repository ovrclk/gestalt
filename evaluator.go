package gestalt

import (
	"context"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/ovrclk/gestalt/result"
	"github.com/ovrclk/gestalt/vars"
)

type Builder interface {
}
type Runable func(Evaluator) result.Result

func Run(node Component) error {
	return NewEvaluator().Evaluate(node).Wait().Err()
}

type Evaluator interface {
	Log() logrus.FieldLogger
	Evaluate(Component) result.Result

	Emit(string, string)
	Vars() vars.Vars

	Context() context.Context
	Builder() Builder
	Stop()
}

type evaluator struct {
	path   string
	ctx    context.Context
	cancel context.CancelFunc
	log    logrus.FieldLogger

	vars vars.Vars
}

func NewEvaluator() *evaluator {
	ctx, cancel := context.WithCancel(context.TODO())
	return &evaluator{
		path:   "",
		ctx:    ctx,
		cancel: cancel,
		log:    logrus.StandardLogger(),
		vars:   vars.NewVars(),
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

func (e *evaluator) Emit(key string, value string) {
	e.vars.Put(key, value)
}

func (e *evaluator) Vars() vars.Vars {
	return e.vars
}

func (e *evaluator) Stop() {
	e.cancel()
}

func (e *evaluator) Evaluate(node Component) result.Result {
	child := e.cloneFor(node)

	child.Log().Debug("start")

	result := node.Eval(child)

	if result.IsError() {
		child.Log().WithError(result.Err()).Error("eval failed")
	}

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
		vars:   e.vars,
	}
}
