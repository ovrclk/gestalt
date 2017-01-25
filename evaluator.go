package gestalt

import (
	"context"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/ovrclk/gestalt/result"
	"github.com/ovrclk/gestalt/vars"
)

type Action func(Evaluator) result.Result

type Evaluator interface {
	Log() logrus.FieldLogger

	Evaluate(Component) result.Result
	Fork(Component) result.Result

	Emit(string, string)
	Vars() vars.Vars

	Message(string, ...interface{})

	Context() context.Context

	Stop()
	Wait()

	HasError() bool
	ClearError()
	Errors() []error
}

type evaluator struct {
	path   string
	ctx    context.Context
	cancel context.CancelFunc
	logger Logger

	vars vars.Vars

	errors []error

	wg       sync.WaitGroup
	children []*evaluator
}

func NewEvaluator() *evaluator {
	return NewEvaluatorWithLogger(newLogBuilder().Logger())
}

func NewEvaluatorWithLogger(logger Logger) *evaluator {
	ctx, cancel := context.WithCancel(context.TODO())
	return &evaluator{
		path:   "",
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
		vars:   vars.NewVars(),
	}
}

func (e *evaluator) Log() logrus.FieldLogger {
	return e.logger.Log()
}

func (e *evaluator) Message(msg string, args ...interface{}) {
	e.logger.Message(msg, args...)
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

func (e *evaluator) Wait() {
	e.wg.Wait()
}

func (e *evaluator) HasError() bool {
	return len(e.errors) > 0
}

func (e *evaluator) ClearError() {
	e.errors = make([]error, 0)
}

func (e *evaluator) Errors() []error {
	return e.errors
}

func (e *evaluator) Evaluate(node Component) result.Result {
	child := e.cloneFor(node)

	m := node.Meta()
	vars.ImportTo(m, e.vars, child.vars)
	e.tracePreEval(child, node)

	result := child.doEvaluate(node)

	vars.ExportTo(m, child.vars, e.vars)

	e.errors = append(e.errors, child.errors...)

	e.tracePostEval(child, node)

	return result
}

func (e *evaluator) doEvaluate(node Component) result.Result {

	e.logger.Start()

	result := node.Eval(e)

	if result.IsError() {
		e.addError(result.Err())
	}

	e.logger.Stop(result.Err())

	return result
}

func (e *evaluator) addError(err error) {
	e.Log().WithError(err).Error("eval failed")
	e.errors = append(e.errors, NewError(e.path, err))
}

func (e *evaluator) Fork(node Component) result.Result {
	e.wg.Add(1)
	go func(child Evaluator) {
		defer e.wg.Done()
		child.Evaluate(node)
		child.Wait()
	}(e.forkFor(node))
	return result.Complete()
}

func (e *evaluator) cloneFor(node Component) *evaluator {
	return e.cloneWithPath(pushPath(e.path, node), node)
}

func (e *evaluator) forkFor(node Component) *evaluator {
	return e.cloneWithPath(e.path, node)
}

func (e *evaluator) cloneWithPath(path string, node Component) *evaluator {
	ctx, cancel := context.WithCancel(e.ctx)

	return &evaluator{
		path:   path,
		ctx:    ctx,
		cancel: cancel,
		logger: e.logger.CloneFor(adornPath(path, node)),
		vars:   vars.NewVars(),
	}
}

func (e *evaluator) tracePreEval(child *evaluator, node Component) {
	//e.Log().Debugf("pre-eval: parent.vars: %v child.vars: %v meta: %v", e.vars, child.vars, node)
}

func (e *evaluator) tracePostEval(child *evaluator, node Component) {
	//e.Log().Debugf("post-eval: parent.vars: %v child.vars: %v meta: %v", e.vars, child.vars, node)
}
