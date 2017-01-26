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

	wg sync.WaitGroup

	pauseOnErr bool
}

func NewEvaluator() *evaluator {
	return NewEvaluatorWithLogger(newLogBuilder().Logger())
}

func NewEvaluatorWithLogger(logger Logger) *evaluator {
	ctx, cancel := context.WithCancel(context.TODO())
	return &evaluator{
		path:       "",
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger,
		vars:       vars.NewVars(),
		pauseOnErr: false,
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
	child := e.cloneWithPath(e.path, node)
	child.pauseOnErr = false
	return child
}

func (e *evaluator) cloneWithPath(path string, node Component) *evaluator {
	ctx, cancel := context.WithCancel(e.ctx)

	return &evaluator{
		path:       path,
		ctx:        ctx,
		cancel:     cancel,
		logger:     e.logger.CloneFor(adornPath(path, node)),
		vars:       vars.NewVars(),
		pauseOnErr: e.pauseOnErr,
	}
}

func (e *evaluator) doPause(err error) bool {
	fmt.Fprintf(os.Stderr, "Error during %v\n", e.path)
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

func (e *evaluator) tracePreEval(child *evaluator, node Component) {
	//e.Log().Debugf("pre-eval: parent.vars: %v child.vars: %v meta: %v", e.vars, child.vars, node)
}

func (e *evaluator) tracePostEval(child *evaluator, node Component) {
	//e.Log().Debugf("post-eval: parent.vars: %v child.vars: %v meta: %v", e.vars, child.vars, node)
}
