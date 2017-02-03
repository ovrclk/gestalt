package exec

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/vars"
)

type CmdFn func(*bufio.Reader, gestalt.Evaluator) error

type Cmd interface {
	gestalt.Component
	FN(CmdFn) Cmd
	WorkingDir(string) Cmd
}

type CC struct {
	cmp  gestalt.Component
	Path string
	Dir  string
	Args []string
	Env  []string

	fn CmdFn
}

func NewCmd(name string, path string, args []string) Cmd {

	requires := vars.Extract(path)
	for _, arg := range args {
		requires = append(requires, vars.Extract(arg)...)
	}

	cmp := gestalt.NewComponent(name, nil).
		WithMeta(vars.NewMeta().Require(requires...))

	return &CC{
		cmp:  cmp,
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

func (c *CC) WorkingDir(dir string) Cmd {
	c.Dir = dir
	return c
}

func (c *CC) Eval(e gestalt.Evaluator) error {

	path := vars.Expand(e.Vars(), c.Path)
	args := make([]string, len(c.Args))

	for i, v := range c.Args {
		args[i] = vars.Expand(e.Vars(), v)
	}

	cmd := exec.CommandContext(e.Context(), path, args...)

	cmd.Dir = vars.Expand(e.Vars(), c.Dir)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	e.Message("running %v %v", path, strings.Join(args, " "))

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("can't execute %v: %v", path, err)
	}

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		logStream(stdoutPipe, e.Log().Debug, stdoutBuf)
	}()
	go func() {
		defer wg.Done()
		logStream(stderrPipe, e.Log().Error, stderrBuf)
	}()
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		if !expectedExecError(err, e) {
			return newError(err, path, args, stdoutBuf, stderrBuf)
		}
	}

	if c.copyStdout() {
		buf := bytes.NewBuffer(stdoutBuf.Bytes())
		err := c.fn(bufio.NewReader(buf), e)
		if err != nil {
			return newError(err, path, args, stdoutBuf, stderrBuf)
		}
		return nil
	}
	return nil
}

func (c *CC) copyStdout() bool {
	return c.fn != nil
}

func logStream(reader io.ReadCloser, log func(...interface{}), b *bytes.Buffer) {
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

func newError(err error, path string, args []string, stdout *bytes.Buffer, stderr *bytes.Buffer) error {
	return &Error{err.Error(), path, args, stdout.String(), stderr.String()}
}
