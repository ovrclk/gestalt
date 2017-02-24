package exec

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/vars"
)

type Pipeline interface {
	Capture(...string) CmdFn
	CaptureAll() CmdFn

	GrepField(string, string) Pipeline
	GrepWith(PipeFilter) Pipeline

	EnsureCount(int) Pipeline
	EnsureWith(PipeValidator) Pipeline
}

type PipeObject map[string]string
type PipeStage func([]PipeObject, gestalt.Evaluator) ([]PipeObject, error)
type PipeParser func(*bufio.Reader, gestalt.Evaluator) ([]PipeObject, error)
type PipeFilter func(PipeObject, gestalt.Evaluator) bool
type PipeValidator func([]PipeObject) error

type pipeline struct {
	pipe    []PipeStage
	parsefn PipeParser
}

func NewObjectPipe(fn PipeParser) Pipeline {
	return &pipeline{make([]PipeStage, 0), fn}
}

func ParseColumns(columns ...string) Pipeline {
	return LineParser(func(line string, _ gestalt.Evaluator) (PipeObject, error) {
		fields := strings.Fields(line)
		obj := make(PipeObject)
		for i := 0; i < len(columns) && i < len(fields); i++ {
			obj[columns[i]] = fields[i]
		}

		return obj, nil
	})
}

func LineParser(fn func(string, gestalt.Evaluator) (PipeObject, error)) Pipeline {
	return NewObjectPipe(func(r *bufio.Reader, e gestalt.Evaluator) ([]PipeObject, error) {
		results := make([]PipeObject, 0)

		scanner := bufio.NewScanner(r)

		for scanner.Scan() {
			line := scanner.Text()

			obj, err := fn(line, e)
			if err != nil {
				return results, err
			}
			results = append(results, obj)
		}

		return results, scanner.Err()
	})
}

func (p *pipeline) EnsureCount(count int) Pipeline {
	return p.EnsureWith(func(objs []PipeObject) error {
		if len(objs) != count {
			return fmt.Errorf("invalid count have:%v want:%v", len(objs), count)
		}
		return nil
	})
}

func (p *pipeline) EnsureWith(fn PipeValidator) Pipeline {
	return p.Then(func(objs []PipeObject, _ gestalt.Evaluator) ([]PipeObject, error) {
		if err := fn(objs); err != nil {
			return objs, err
		}
		return objs, nil
	})
}

func (p *pipeline) GrepField(key string, value string) Pipeline {
	return p.GrepWith(func(obj PipeObject, e gestalt.Evaluator) bool {
		if v, ok := obj[key]; ok {
			expanded := vars.Expand(e.Vars(), value)
			return v == expanded
		}
		return false
	})
}

func (p *pipeline) GrepWith(fn PipeFilter) Pipeline {
	return p.Then(func(objs []PipeObject, e gestalt.Evaluator) ([]PipeObject, error) {
		result := make([]PipeObject, 0)
		for _, obj := range objs {
			if fn(obj, e) {
				result = append(result, obj)
			}
		}
		return result, nil
	})
}

func (p *pipeline) Capture(keys ...string) CmdFn {
	return p.finally(func(objs []PipeObject, e gestalt.Evaluator) error {
		for _, obj := range objs {
			for _, k := range keys {
				if v, ok := obj[k]; ok {
					e.Emit(k, v)
				}
			}
		}
		return nil
	})
}

func (p *pipeline) CaptureAll() CmdFn {
	return p.finally(func(objs []PipeObject, e gestalt.Evaluator) error {
		for _, obj := range objs {
			for k, v := range obj {
				e.Emit(k, v)
			}
		}
		return nil
	})
}

func (p *pipeline) Then(fn PipeStage) *pipeline {
	p.pipe = append(p.pipe, fn)
	return p
}

func (p *pipeline) finally(fn func(objs []PipeObject, e gestalt.Evaluator) error) CmdFn {
	return func(r *bufio.Reader, e gestalt.Evaluator) error {
		objs, err := p.process(r, e)
		if err != nil {
			return err
		}
		return fn(objs, e)
	}
}

func (p *pipeline) process(r *bufio.Reader, e gestalt.Evaluator) ([]PipeObject, error) {
	objs, err := p.parsefn(r, e)
	if err != nil {
		return objs, err
	}

	for _, fn := range p.pipe {
		objs, err = fn(objs, e)
		if err != nil {
			return objs, err
		}
	}
	return objs, nil
}
