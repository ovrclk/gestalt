package gestalt

import (
	"context"

	"github.com/ovrclk/gestalt/vars"
)

type path struct {
	base string
	name string
}

type pathHandler struct {
	stack []path
}

func newPathHandler() *pathHandler {
	return &pathHandler{[]path{path{}}}
}

func (h *pathHandler) Push(_ Evaluator, node Component) {
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

func (h *pathHandler) Pop(_ Evaluator, _ Component) {
	if sz := len(h.stack); sz > 0 {
		h.stack = h.stack[0 : sz-1]
	}
}

func (h *pathHandler) Clone() *pathHandler {
	return &pathHandler{[]path{h.top()}}
}

func (h *pathHandler) Current() string {
	return h.top().name
}

func (h *pathHandler) Base() string {
	return h.top().base
}

func (h *pathHandler) top() path {
	return h.stack[len(h.stack)-1]
}

type logHandler struct {
	stack []Logger
}

func newLogHandler(l Logger) *logHandler {
	return &logHandler{[]Logger{l}}
}

func (h *logHandler) Push(e Evaluator, _ Component) {
	h.stack = append(h.stack, h.Current().CloneFor(e.Path()))
}

func (h *logHandler) Pop(_ Evaluator, _ Component) {
	if sz := len(h.stack); sz > 1 {
		h.stack = h.stack[0 : sz-1]
	}
}

func (h *logHandler) Clone() *logHandler {
	top := h.Current().Clone()
	return &logHandler{[]Logger{top}}
}

func (h *logHandler) Current() Logger {
	return h.stack[len(h.stack)-1]
}

type ctxState struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type ctxHandler struct {
	stack []*ctxState
}

func newCtxHandler() *ctxHandler {
	ctx, cancel := context.WithCancel(context.TODO())
	top := &ctxState{ctx, cancel}
	return &ctxHandler{[]*ctxState{top}}
}

func (h *ctxHandler) Push(_ Evaluator, _ Component) {
	ctx, cancel := context.WithCancel(h.Current())
	h.stack = append(h.stack, &ctxState{ctx, cancel})
}

func (h *ctxHandler) Pop(_ Evaluator, _ Component) {
	if sz := len(h.stack); sz > 1 {
		h.stack = h.stack[0 : sz-1]
	}
}

func (h *ctxHandler) Clone() *ctxHandler {
	ctx, cancel := context.WithCancel(h.Current())
	top := &ctxState{ctx, cancel}
	return &ctxHandler{[]*ctxState{top}}
}

func (h *ctxHandler) Current() context.Context {
	return h.stack[len(h.stack)-1].ctx
}

func (h *ctxHandler) Cancel() {
	h.stack[len(h.stack)-1].cancel()
}

type varHandler struct {
	stack []vars.Vars
}

func newVarHandler() *varHandler {
	return &varHandler{[]vars.Vars{vars.NewVars()}}
}

func (h *varHandler) Push(_ Evaluator, node Component) {
	new := vars.NewVars()
	vars.ImportTo(node.Meta(), h.Current(), new)
	h.stack = append(h.stack, new)
}

func (h *varHandler) Pop(_ Evaluator, node Component) {
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

func (h *varHandler) Clone() *varHandler {
	top := vars.NewVars().Merge(h.Current())
	return &varHandler{[]vars.Vars{top}}
}

func (h *varHandler) Current() vars.Vars {
	if sz := len(h.stack); sz > 0 {
		return h.stack[sz-1]
	} else {
		return vars.NewVars()
	}
}

type traceHandler struct {
}

func newTraceHandler() *traceHandler {
	return &traceHandler{}
}

func (h *traceHandler) Push(e Evaluator, node Component) {
}

func (h *traceHandler) Pop(e Evaluator, node Component) {
}

func (h *traceHandler) Clone() *traceHandler {
	return &traceHandler{}
}
