package gestalt

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/ovrclk/gestalt/vars"
)

type path struct {
	base string
	name string
}

type pathVisitor struct {
	stack []path
}

func newPathVisitor() *pathVisitor {
	return &pathVisitor{[]path{path{}}}
}

func (h *pathVisitor) Push(_ Traverser, node Component) {
	var top path

	base := h.Base()
	next := base + "/" + node.Name()

	if node.IsPassThrough() {
		top = path{base, next}
	} else {
		top = path{next, next}
	}

	h.stack = append(h.stack, top)
}

func (h *pathVisitor) Pop(_ Traverser, _ Component) {
	if sz := len(h.stack); sz > 0 {
		h.stack = h.stack[0 : sz-1]
	}
}

func (h *pathVisitor) Clone() *pathVisitor {
	return &pathVisitor{[]path{h.top()}}
}

func (h *pathVisitor) Current() string {
	return h.top().name
}

func (h *pathVisitor) Base() string {
	return h.top().base
}

func (h *pathVisitor) top() path {
	return h.stack[len(h.stack)-1]
}

type logVisitor struct {
	stack []Logger
}

func newLogVisitor(l Logger) *logVisitor {
	return &logVisitor{[]Logger{l}}
}

func (h *logVisitor) Push(e Traverser, _ Component) {
	h.stack = append(h.stack, h.Current().CloneFor(e.Path()))
}

func (h *logVisitor) Pop(_ Traverser, _ Component) {
	if sz := len(h.stack); sz > 1 {
		h.stack = h.stack[0 : sz-1]
	}
}

func (h *logVisitor) Clone() *logVisitor {
	top := h.Current().Clone()
	return &logVisitor{[]Logger{top}}
}

func (h *logVisitor) Current() Logger {
	return h.stack[len(h.stack)-1]
}

type ctxState struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type ctxVisitor struct {
	stack []*ctxState
}

func newCtxVisitor() *ctxVisitor {
	ctx, cancel := context.WithCancel(context.TODO())
	top := &ctxState{ctx, cancel}
	return &ctxVisitor{[]*ctxState{top}}
}

func (h *ctxVisitor) Push(_ Traverser, _ Component) {
	ctx, cancel := context.WithCancel(h.Current())
	h.stack = append(h.stack, &ctxState{ctx, cancel})
}

func (h *ctxVisitor) Pop(_ Traverser, _ Component) {
	if sz := len(h.stack); sz > 1 {
		h.stack = h.stack[0 : sz-1]
	}
}

func (h *ctxVisitor) Clone() *ctxVisitor {
	ctx, cancel := context.WithCancel(h.Current())
	top := &ctxState{ctx, cancel}
	return &ctxVisitor{[]*ctxState{top}}
}

func (h *ctxVisitor) Current() context.Context {
	return h.stack[len(h.stack)-1].ctx
}

func (h *ctxVisitor) Cancel() {
	h.stack[len(h.stack)-1].cancel()
}

type varVisitor struct {
	stack []vars.Vars
}

func newVarVisitor() *varVisitor {
	return &varVisitor{[]vars.Vars{vars.NewVars()}}
}

func (h *varVisitor) Push(_ Traverser, node Component) {
	new := vars.NewVars()
	vars.ImportTo(node.Meta(), h.Current(), new)
	h.stack = append(h.stack, new)
}

func (h *varVisitor) Pop(_ Traverser, node Component) {
	sz := len(h.stack)
	switch {
	case sz > 1:
		top := h.stack[sz-1]
		new := h.stack[sz-2]
		vars.ExportTo(node.Meta(), top, new)
		fallthrough
	case sz > 0:
		h.stack = h.stack[0 : sz-1]
	}
}

func (h *varVisitor) Clone() *varVisitor {
	top := vars.NewVars().Merge(h.Current())
	return &varVisitor{[]vars.Vars{top}}
}

func (h *varVisitor) Current() vars.Vars {
	if sz := len(h.stack); sz > 0 {
		return h.stack[sz-1]
	} else {
		return vars.NewVars()
	}
}

type errVisitor struct {
	stack [][]error
}

func newErrVisitor() *errVisitor {
	return &errVisitor{[][]error{[]error{}}}
}

func (h *errVisitor) Push(_ Traverser, _ Component) {
	h.stack = append(h.stack, []error{})
}

func (h *errVisitor) Pop(_ Traverser, _ Component) {
	top := h.stack[len(h.stack)-1]
	next := append(h.stack[len(h.stack)-2], top...)
	h.stack = h.stack[0 : len(h.stack)-1]
	h.stack[len(h.stack)-1] = next
}

func (h *errVisitor) Clone() *errVisitor {
	return newErrVisitor()
}

func (h *errVisitor) Current() []error {
	return h.stack[len(h.stack)-1]
}

func (h *errVisitor) Clear() {
	h.stack[len(h.stack)-1] = []error{}
}

func (h *errVisitor) Add(err error) {
	h.stack[len(h.stack)-1] = append(h.stack[len(h.stack)-1], err)
}

type waitState struct {
	wg       *sync.WaitGroup
	children []*sync.WaitGroup
}

type waitVisitor struct {
	stack []*waitState
}

func newWaitVisitor() *waitVisitor {
	return &waitVisitor{[]*waitState{&waitState{}}}
}

func (h *waitVisitor) Push(_ Traverser, _ Component) {
	h.stack = append(h.stack, &waitState{})
}

func (h *waitVisitor) Pop(_ Traverser, _ Component) {
	top := h.stack[len(h.stack)-1]
	next := h.stack[len(h.stack)-2]

	if top.wg != nil {
		next.children = append(next.children, top.wg)
	}
	next.children = append(next.children, top.children...)

	h.stack = h.stack[0 : len(h.stack)-1]
}

func (h *waitVisitor) Clone() *waitVisitor {
	return newWaitVisitor()
}

func (h *waitVisitor) Current() *sync.WaitGroup {
	top := h.stack[len(h.stack)-1]
	if top.wg == nil {
		top.wg = new(sync.WaitGroup)
	}
	return top.wg
}

func (h *waitVisitor) Wait() {
	top := h.stack[len(h.stack)-1]
	if top.wg != nil {
		top.wg.Wait()
	}
	for _, child := range top.children {
		child.Wait()
	}
}

type nodeVisitor struct {
	stack []Component
}

func newNodeVisitor() *nodeVisitor {
	return &nodeVisitor{}
}

func (h *nodeVisitor) Push(_ Traverser, node Component) {
	h.stack = append(h.stack, node)
}

func (h *nodeVisitor) Pop(_ Traverser, _ Component) {
	h.stack = h.stack[0 : len(h.stack)-1]
}

func (h *nodeVisitor) Clone() *nodeVisitor {
	return newNodeVisitor()
}

func (h *nodeVisitor) Root() Component {
	return h.stack[0]
}

type traceVisitor struct {
	out io.Writer
}

func newTraceVisitor(out io.Writer) *traceVisitor {
	return &traceVisitor{out}
}

func (h *traceVisitor) Push(t Traverser, node Component) {
	fmt.Fprintf(h.out, "TRACE ENTER [%v] [%v]\n", t.Path(), node.Name())
	if e, ok := t.(Evaluator); ok {
		fmt.Fprintf(h.out, "ERRORS: %v\n", e.Errors())
		fmt.Fprintf(h.out, "VARS: %v\n", e.Vars())
	}
}

func (h *traceVisitor) Pop(t Traverser, node Component) {
	fmt.Fprintf(h.out, "TRACE LEAVE [%v] [%v]\n", t.Path(), node.Name())
	if e, ok := t.(Evaluator); ok {
		fmt.Fprintf(h.out, "ERRORS: %v\n", e.Errors())
		fmt.Fprintf(h.out, "VARS: %v\n", e.Vars())
	}
}

func (h *traceVisitor) Clone() *traceVisitor {
	return &traceVisitor{}
}
