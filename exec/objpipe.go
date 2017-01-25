package exec

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/ovrclk/gestalt"
)

type ObjectPipe interface {
	Capture(...string) CmdFn
	CaptureAll() CmdFn

	GrepField(string, string) ObjectPipe
	GrepWith(PipeObjGrepFn) ObjectPipe

	EnsureCount(int) ObjectPipe
	EnsureWith(PipeObjEnsureFn) ObjectPipe
}

type PipeObject map[string]string
type PipeObjFn func([]PipeObject, gestalt.Evaluator) ([]PipeObject, error)
type PipeObjParseFn func(*bufio.Reader, gestalt.Evaluator) ([]PipeObject, error)
type PipeObjGrepFn func(PipeObject, gestalt.Evaluator) bool
type PipeObjEnsureFn func([]PipeObject) error

type objpipe struct {
	pipe    []PipeObjFn
	parsefn PipeObjParseFn
}

func NewObjectPipe(fn PipeObjParseFn) ObjectPipe {
	return &objpipe{make([]PipeObjFn, 0), fn}
}

func ParseColumns(columns ...string) ObjectPipe {
	return LineParser(func(line string, _ gestalt.Evaluator) (PipeObject, error) {
		fields := strings.Fields(line)
		obj := make(PipeObject)
		for i := 0; i < len(columns) && i < len(fields); i++ {
			obj[columns[i]] = fields[i]
		}

		return obj, nil
	})
}

func LineParser(fn func(string, gestalt.Evaluator) (PipeObject, error)) ObjectPipe {
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

func (p *objpipe) EnsureCount(count int) ObjectPipe {
	return p.EnsureWith(func(objs []PipeObject) error {
		if len(objs) != count {
			return fmt.Errorf("invalid count have:%v want:%v", len(objs), count)
		}
		return nil
	})
}

func (p *objpipe) EnsureWith(fn PipeObjEnsureFn) ObjectPipe {
	return p.Then(func(objs []PipeObject, _ gestalt.Evaluator) ([]PipeObject, error) {
		if err := fn(objs); err != nil {
			return objs, err
		}
		return objs, nil
	})
}

func (p *objpipe) GrepField(key string, value string) ObjectPipe {
	return p.GrepWith(func(obj PipeObject, e gestalt.Evaluator) bool {
		if v, ok := obj[key]; ok {
			return v == value
		}
		return false
	})
}

func (p *objpipe) GrepWith(fn PipeObjGrepFn) ObjectPipe {
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

func (p *objpipe) Capture(keys ...string) CmdFn {
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

func (p *objpipe) CaptureAll() CmdFn {
	return p.finally(func(objs []PipeObject, e gestalt.Evaluator) error {
		for _, obj := range objs {
			for k, v := range obj {
				e.Emit(k, v)
			}
		}
		return nil
	})
}

func (p *objpipe) Then(fn PipeObjFn) *objpipe {
	p.pipe = append(p.pipe, fn)
	return p
}

func (p *objpipe) finally(fn func(objs []PipeObject, e gestalt.Evaluator) error) CmdFn {
	return func(r *bufio.Reader, e gestalt.Evaluator) error {
		objs, err := p.process(r, e)
		fmt.Println(objs)
		if err != nil {
			return err
		}
		return fn(objs, e)
	}
}

func (p *objpipe) process(r *bufio.Reader, e gestalt.Evaluator) ([]PipeObject, error) {
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
