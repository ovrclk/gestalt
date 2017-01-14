package exec

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"strings"
	"syscall"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/result"
)

type CmdFn func(*bufio.Reader, gestalt.Evaluator) error

type Cmd struct {
	gestalt.C
	Path string
	Args []string
	Env  []string

	fn CmdFn
}

func NewCmd(name string, path string, args []string) *Cmd {
	return &Cmd{
		C:    *gestalt.NewComponent(name, nil),
		Path: path,
		Args: args,
	}
}

func (c *Cmd) FN(fn CmdFn) *Cmd {
	c.fn = fn
	return c
}

func (c *Cmd) Eval(e gestalt.Evaluator) result.Result {
	return c.Build(e.Builder())(e)
}

func (c *Cmd) Build(b gestalt.Builder) gestalt.Runable {
	return func(e gestalt.Evaluator) result.Result {
		cmd := exec.CommandContext(e.Context(), c.Path, c.Args...)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return result.Error(err)
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return result.Error(err)
		}

		e.Log().Debugf("running %v %v", cmd.Path, cmd.Args)
		if err := cmd.Start(); err != nil {
			e.Log().WithError(err).Errorf("error running %v", cmd.Path)
			return result.Error(err)
		}

		var buf *bytes.Buffer
		if c.copyStdout() {
			buf = new(bytes.Buffer)
		}

		go logStream(stdout, e.Log().Debug, buf)
		go logStream(stderr, e.Log().Error, nil)

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
}

func (c *Cmd) copyStdout() bool {
	return c.fn != nil
}

func logStream(reader io.ReadCloser, log func(fmt ...interface{}), b *bytes.Buffer) {
	buf := make([]byte, 80)
	for {
		n, err := reader.Read(buf)

		if b != nil && n > 0 {
			b.Write(buf[0:n])
		}

		if err != nil {
			break
		}

		log(strings.TrimRight(string(buf[0:n]), "\n\r"))
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
