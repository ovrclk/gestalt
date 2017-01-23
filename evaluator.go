package gestalt

import (
	"context"
	"strings"

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

	LogStdout(string)
	LogStderr(string)
	LogMessage(string)

	Context() context.Context
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
	return NewEvaluatorWithLogger(logrus.StandardLogger())
}

func NewEvaluatorWithLogger(logger *logrus.Logger) *evaluator {
	ctx, cancel := context.WithCancel(context.TODO())
	return &evaluator{
		path:   "",
		ctx:    ctx,
		cancel: cancel,
		log:    logger,
		vars:   vars.NewVars(),
	}
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

	m := node.Meta()
	vars.ImportTo(m, e.vars, child.vars)
	e.tracePreEval(child, node)

	result := child.doEvaluate(node)

	vars.ExportTo(m, child.vars, e.vars)
	e.tracePostEval(child, node)

	return result
}

func (e *evaluator) doEvaluate(node Component) result.Result {
	e.Log().Debug("start")

	result := node.Eval(e)

	if result.IsError() {
		e.Log().WithError(result.Err()).Error("eval failed")
	}

	e.Log().Debugf("end -> %v", result)

	return result
}

func (e *evaluator) Fork(node Component) result.Result {
	ch := make(chan result.Result)
	go func(e Evaluator) {
		defer close(ch)
		ch <- e.Evaluate(node).Wait()
	}(e.forkFor(node))
	return result.Running(func() result.Result {
		return <-ch
	})
}

func (e *evaluator) cloneFor(node Component) *evaluator {
	return e.cloneWithPath(pushPath(e.path, node))
}

func (e *evaluator) forkFor(node Component) *evaluator {
	return e.cloneWithPath(e.path)
}

func (e *evaluator) cloneWithPath(path string) *evaluator {
	ctx, cancel := context.WithCancel(e.ctx)
	return &evaluator{
		path:   path,
		ctx:    ctx,
		cancel: cancel,
		log:    e.log.WithField("path", path),
		vars:   vars.NewVars(),
	}
}

func (e *evaluator) LogStderr(buf string) {
	e.Log().Error(strings.TrimRight(buf, "\n\r"))
}

func (e *evaluator) LogStdout(buf string) {
	e.Log().Debug(strings.TrimRight(buf, "\n\r"))
}

func (e *evaluator) LogMessage(string) {
}

func (e *evaluator) tracePreEval(child *evaluator, node Component) {
	//e.Log().Debugf("pre-eval: parent.vars: %v child.vars: %v meta: %v", e.vars, child.vars, node)
}

func (e *evaluator) tracePostEval(child *evaluator, node Component) {
	//e.Log().Debugf("post-eval: parent.vars: %v child.vars: %v meta: %v", e.vars, child.vars, node)
}
