package gestalt

import (
	"context"

	"github.com/Sirupsen/logrus"
	"github.com/ovrclk/gestalt/result"
	"github.com/ovrclk/gestalt/vars"
)

type Evaluator interface {
	Log() logrus.FieldLogger
	Path() string

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

	Root() Component
}

type evaluator struct {
	node *nodeVisitor
	path *pathVisitor
	log  *logVisitor
	vars *varVisitor
	ctx  *ctxVisitor
	err  *errVisitor
	wait *waitVisitor

	visitors []Visitor

	handler evalHandler
}

type evalHandler interface {
	Eval(Evaluator, Component) result.Result
}

func NewEvaluator(visitors ...Visitor) *evaluator {
	return NewEvaluatorWithLogger(newLogBuilder().Logger(), visitors...)
}

func NewEvaluatorWithLogger(logger Logger, visitors ...Visitor) *evaluator {
	return &evaluator{
		node:     newNodeVisitor(),
		path:     newPathVisitor(),
		log:      newLogVisitor(logger),
		vars:     newVarVisitor(),
		ctx:      newCtxVisitor(),
		err:      newErrVisitor(),
		wait:     newWaitVisitor(),
		visitors: visitors,
		handler:  defaultEvalHandler,
	}
}

func (e *evaluator) Log() logrus.FieldLogger {
	return e.log.Current().Log()
}

func (e *evaluator) Path() string {
	return e.path.Current()
}

func (e *evaluator) Message(msg string, args ...interface{}) {
	e.log.Current().Message(msg, args...)
}

func (e *evaluator) Context() context.Context {
	return e.ctx.Current()
}

func (e *evaluator) Emit(key string, value string) {
	e.vars.Current().Put(key, value)
}

func (e *evaluator) Vars() vars.Vars {
	return e.vars.Current()
}

func (e *evaluator) Stop() {
	e.ctx.Cancel()
}

func (e *evaluator) Wait() {
	e.wait.Wait()
}

func (e *evaluator) HasError() bool {
	return len(e.err.Current()) > 0
}

func (e *evaluator) ClearError() {
	e.err.Clear()
}

func (e *evaluator) Errors() []error {
	return e.err.Current()
}

func (e *evaluator) Evaluate(node Component) result.Result {
	e.push(node)

	result := e.handler.Eval(e, node)

	if result.IsError() {
		e.addError(result.Err())
	}

	e.pop(node)
	return result
}

func (e *evaluator) push(node Component) {
	e.node.Push(e, node)
	e.path.Push(e, node)
	e.log.Push(e, node)
	e.ctx.Push(e, node)
	e.vars.Push(e, node)
	e.err.Push(e, node)
	e.wait.Push(e, node)

	for _, v := range e.visitors {
		v.Push(e, node)
	}
}

func (e *evaluator) pop(node Component) {

	for i := len(e.visitors) - 1; i >= 0; i-- {
		e.visitors[i].Pop(e, node)
	}

	e.wait.Pop(e, node)
	e.err.Pop(e, node)
	e.vars.Pop(e, node)
	e.log.Pop(e, node)
	e.ctx.Pop(e, node)
	e.path.Pop(e, node)
	e.node.Pop(e, node)
}

func (e *evaluator) addError(err error) {
	e.Log().WithError(err).Error("eval failed")
	e.err.Add(NewError(e.Path(), err))
}

func (e *evaluator) Fork(node Component) result.Result {
	wg := e.wait.Current()
	wg.Add(1)
	go func(child Evaluator) {
		defer wg.Done()
		child.Evaluate(node)
		child.Wait()
	}(e.forkFor(node))
	return result.Complete()
}

func (e *evaluator) Root() Component {
	return e.node.Root()
}

func (e *evaluator) forkFor(node Component) *evaluator {
	return &evaluator{
		node:    e.node.Clone(),
		path:    e.path.Clone(),
		log:     e.log.Clone(),
		vars:    e.vars.Clone(),
		ctx:     e.ctx.Clone(),
		err:     e.err.Clone(),
		wait:    e.wait.Clone(),
		handler: defaultEvalHandler,
	}
}

type _defaultEvalHandler struct{}

var defaultEvalHandler = _defaultEvalHandler{}

func (eh _defaultEvalHandler) Eval(e Evaluator, node Component) result.Result {
	return node.Eval(e)
}
