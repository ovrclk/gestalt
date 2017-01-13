package exec

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"strings"
	"syscall"

	"github.com/ovrclk/gestalt"
)

type CmdFn func(*bufio.Reader, gestalt.RunCtx) (gestalt.ResultValues, error)

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

func (c *Cmd) Exports(names ...string) gestalt.Component {
	c.C.Exports(names...)
	return c
}
func (c *Cmd) Imports(names ...string) gestalt.Component {
	c.C.Imports(names...)
	return c
}
func (c *Cmd) Requires(names ...string) gestalt.Component {
	c.C.Requires(names...)
	return c
}

func (c *Cmd) Build(bctx gestalt.BuildCtx) gestalt.Runable {
	return func(rctx gestalt.RunCtx) gestalt.Result {
		cmd := exec.CommandContext(rctx.Context(), c.Path, c.Args...)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return gestalt.ResultError(err)
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return gestalt.ResultError(err)
		}

		rctx.Logger().Debugf("running %v %v", cmd.Path, cmd.Args)
		if err := cmd.Start(); err != nil {
			rctx.Logger().WithError(err).Errorf("error running %v", cmd.Path)
			return gestalt.ResultError(err)
		}

		var buf *bytes.Buffer
		if c.copyStdout() {
			buf = new(bytes.Buffer)
		}

		go logStream(stdout, rctx.Logger().Debug, buf)
		go logStream(stderr, rctx.Logger().Error, nil)

		if err := cmd.Wait(); err != nil {
			if !expectedExecError(err, rctx) {
				rctx.Logger().WithError(err).Error("command failed")
				return gestalt.ResultError(err)
			}
		}

		if c.copyStdout() {
			vals, err := c.fn(bufio.NewReader(buf), rctx)
			if err != nil {
				return gestalt.NewResult(gestalt.RunStateError, vals, err)
			}
			return gestalt.NewResult(gestalt.RunStateComplete, vals, nil)
		}

		return gestalt.ResultSuccess()
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
func expectedExecError(err error, rctx gestalt.RunCtx) bool {
	if exiterr, ok := err.(*exec.ExitError); ok && rctx.Context().Err() != nil {
		if wstatus, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			if wstatus.Signal() == syscall.SIGKILL {
				return true
			}
		}
	}
	return false
}
