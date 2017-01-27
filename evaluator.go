package gestalt

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/ovrclk/gestalt/result"
	"github.com/ovrclk/gestalt/vars"
)

type Action func(Evaluator) result.Result

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
}

type evaluator struct {
	path *pathVisitor
	log  *logVisitor
	vars *varVisitor
	ctx  *ctxVisitor

	errors []error

	wg sync.WaitGroup

	pauseOnErr bool
}

func NewEvaluator() *evaluator {
	return NewEvaluatorWithLogger(newLogBuilder().Logger())
}

func NewEvaluatorWithLogger(logger Logger) *evaluator {
	return &evaluator{
		path:       newPathVisitor(),
		log:        newLogVisitor(logger),
		vars:       newVarVisitor(),
		ctx:        newCtxVisitor(),
		pauseOnErr: false,
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

	e.push(node)

	result := e.doEvaluate(node)

	if result.IsError() {
		e.addError(result.Err())
	}

	e.pop(node)

	return result
}

func (e *evaluator) push(node Component) {
	e.path.Push(e, node)
	e.log.Push(e, node)
	e.ctx.Push(e, node)
	e.vars.Push(e, node)
}

func (e *evaluator) pop(node Component) {
	e.vars.Pop(e, node)
	e.log.Pop(e, node)
	e.ctx.Pop(e, node)
	e.path.Pop(e, node)
}

func (e *evaluator) doEvaluate(node Component) result.Result {

	var result result.Result

	for {
		result = node.Eval(e)

		if !result.IsError() {
			break
		}

		if !e.pauseOnErr {
			break
		}

		if e.doPause(result.Err()) {
			break
		}
	}

	return result
}

func (e *evaluator) addError(err error) {
	e.Log().WithError(err).Error("eval failed")
	e.errors = append(e.errors, NewError(e.Path(), err))
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

func (e *evaluator) forkFor(node Component) *evaluator {
	return &evaluator{
		path:       e.path.Clone(),
		log:        e.log.Clone(),
		vars:       e.vars.Clone(),
		ctx:        e.ctx.Clone(),
		pauseOnErr: false,
	}
}

func (e *evaluator) doPause(err error) bool {
	fmt.Fprintf(os.Stderr, "Error during %v\n", e.Path())
	fmt.Fprintf(os.Stderr, "%v\n", err)
	if err, ok := err.(ErrorWithDetail); ok {
		fmt.Fprintf(os.Stderr, "%v\n", err.Detail())
	}

	fmt.Fprintf(os.Stderr, "Current Vars:\n")
	for _, k := range e.Vars().Keys() {
		fmt.Fprintf(os.Stderr, "%v=%v\n", k, e.Vars().Get(k))
	}

	fmt.Fprintf(os.Stderr, "Retry? [y/n]:")

	bytes := make([]byte, 200)
	n, err := os.Stdin.Read(bytes)
	if err != nil {
		return true
	}

	if n > 0 && bytes[0] == 'y' {
		return false
	}
	return true
}
