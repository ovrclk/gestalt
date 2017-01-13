package exec

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/ovrclk/gestalt"
)

type TextPipe interface {
	Capture(...string) CmdFn
	Head() TextPipe
}

type textpipe struct {
	pipe []pipefn

	l logrus.FieldLogger
}

type pipefn func(*bufio.Reader, *bufio.Writer) error

func P() *textpipe {
	return &textpipe{}
}

func (f *textpipe) Capture(keys ...string) CmdFn {
	return func(r *bufio.Reader, rctx gestalt.RunCtx) (gestalt.ResultValues, error) {

		f.l = rctx.Logger()

		r, err := f.process(r)
		if err != nil {
			return nil, err
		}

		values := make(gestalt.ResultValues)

		line, err := r.ReadString('\n')
		if err != nil && err != io.EOF {
			return values, err
		}

		fields := strings.Fields(line)

		for i := 0; i < len(keys) && i < len(fields); i++ {
			values[keys[i]] = fields[i]
		}
		return values, nil
	}
}

func (f *textpipe) Head() TextPipe {
	f.compose(func(r *bufio.Reader, w *bufio.Writer) error {

		// copy the first line
		line, err := r.ReadBytes('\n')

		if err != nil {
			return err
		}

		_, err = w.Write(line)
		if err != nil {
			return err
		}

		return nil
	})
	return f
}

func (f *textpipe) compose(fn pipefn) {
	f.pipe = append(f.pipe, fn)
}

func (f *textpipe) process(r *bufio.Reader) (*bufio.Reader, error) {
	var w *bytes.Buffer
	for _, childfn := range f.pipe {
		w = new(bytes.Buffer)
		wb := bufio.NewWriter(w)

		if err := childfn(r, wb); err != nil {
			return r, err
		}

		if err := wb.Flush(); err != nil {
			return r, err
		}

		r = bufio.NewReader(w)
	}
	return r, nil
}
