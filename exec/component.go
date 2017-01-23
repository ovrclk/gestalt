package exec

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/result"
	"github.com/ovrclk/gestalt/vars"
)

type CmdFn func(*bufio.Reader, gestalt.Evaluator) error

type Cmd interface {
	gestalt.Component
	FN(CmdFn) Cmd
}

type CC struct {
	cmp  gestalt.Component
	Path string
	Args []string
	Env  []string

	fn CmdFn
}

func NewCmd(name string, path string, args []string) Cmd {
	return &CC{
		cmp:  gestalt.NewComponent(name, nil),
		Path: path,
		Args: args,
	}
}

func (c *CC) Name() string {
	return c.cmp.Name()
}

func (c *CC) IsPassThrough() bool {
	return false
}

func (c *CC) WithMeta(m vars.Meta) gestalt.Component {
	c.cmp.WithMeta(m)
	return c
}

func (c *CC) Meta() vars.Meta {
	return c.cmp.Meta()
}

func (c *CC) FN(fn CmdFn) Cmd {
	c.fn = fn
	return c
}

func (c *CC) Eval(e gestalt.Evaluator) result.Result {

	path := vars.Expand(e.Vars(), c.Path)
	args := make([]string, len(c.Args))

	for i, v := range c.Args {
		args[i] = vars.Expand(e.Vars(), v)
	}

	cmd := exec.CommandContext(e.Context(), path, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return result.Error(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return result.Error(err)
	}

	e.Log().Debugf("running %v %v", path, strings.Join(args, " "))

	if err := cmd.Start(); err != nil {
		e.Log().WithError(err).Errorf("error running %v", cmd.Path)
		return result.Error(err)
	}

	var buf *bytes.Buffer
	if c.copyStdout() {
		buf = new(bytes.Buffer)
	}

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		logStream(stdout, e.LogStdout, buf)
	}()
	go func() {
		defer wg.Done()
		logStream(stderr, e.LogStderr, nil)
	}()
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		if !expectedExecError(err, e) {
			e.Log().WithError(err).Error("command failed")
			return result.Error(err)
		}
	}

	if c.copyStdout() {
		err := c.fn(bufio.NewReader(buf), e)
		if err != nil {
			return result.Error(err)
		}
		return result.Complete()
	}
	return result.Complete()
}

func (c *CC) copyStdout() bool {
	return c.fn != nil
}

func logStream(reader io.ReadCloser, log func(string), b *bytes.Buffer) {
	buf := make([]byte, 80)
	for {
		n, err := reader.Read(buf)

		if b != nil && n > 0 {
			b.Write(buf[0:n])
		}

		if err != nil {
			break
		}

		if n > 0 {
			log(string(buf[0:n]))
		}
	}
}

// Fail silently if killed due to context being cancelled.
func expectedExecError(err error, e gestalt.Evaluator) bool {
	if exiterr, ok := err.(*exec.ExitError); ok && e.Context().Err() != nil {
		if wstatus, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			if wstatus.Signal() == syscall.SIGKILL {
				return true
			}
		}
	}
	return false
}
