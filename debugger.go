package gestalt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/ovrclk/gestalt/result"
)

type commandResult string

const (
	retryResult    commandResult = "retry"
	continueResult commandResult = "continue"
)

type debugHandler struct {

	// stop before execution.
	breakpoints []string

	// stop after execution if failed.
	failpoints []string

	in  io.Reader
	out io.Writer
}

func newDebugHandler(in io.Reader, out io.Writer) *debugHandler {
	return &debugHandler{in: in, out: out}
}

func (h *debugHandler) AddBreakpoint(expr string) error {
	h.breakpoints = append(h.breakpoints, expr)
	return nil
}

func (h *debugHandler) AddFailpoint(expr string) error {
	h.failpoints = append(h.failpoints, expr)
	return nil
}

func (h *debugHandler) Eval(e Evaluator, node Component) result.Result {

	if h.matchBreakPath(e.Path(), h.breakpoints) {
		h.runBreakConsole(e, node)
	}

	var result result.Result

	for {
		result = node.Eval(e)

		if result.IsComplete() {
			break
		}

		if !h.matchBreakPath(e.Path(), h.failpoints) {
			break
		}

		if h.runFailureConsole(e, node, result.Err()) != retryResult {
			break
		}
	}

	return result
}

func (h *debugHandler) runBreakConsole(e Evaluator, node Component) commandResult {
	for i := 0; ; i++ {
		app := h.makeBreakApp()

		if i == 0 {
			app.app.Usage([]string{})
		}

		cmd, err := h.readCommand(app.app)
		if err == io.EOF {
			return continueResult
		}

		switch cmd {
		case app.cmdContinue.FullCommand():
			fmt.Fprintf(h.out, "continuing...\n")
			return continueResult
		case app.cmdErrors.FullCommand():
			h.showErrors(e, node)
		case app.cmdVars.FullCommand():
			h.showVars(e, node)
		}
	}
	return continueResult
}

func (h *debugHandler) runFailureConsole(e Evaluator, node Component, curerr error) commandResult {
	for i := 0; ; i++ {
		app := h.makeFailureApp()

		if i == 0 {
			app.app.Usage([]string{})
		}

		cmd, err := h.readCommand(app.app)
		if err == io.EOF {
			return continueResult
		}

		switch cmd {
		case app.cmdContinue.FullCommand():
			fmt.Fprintf(h.out, "continuing...\n")
			return continueResult
		case app.cmdRetry.FullCommand():
			fmt.Fprintf(h.out, "retrying...\n")
			return retryResult
		case app.cmdErrors.FullCommand():
			h.showErrors(e, node, curerr)
		case app.cmdVars.FullCommand():
			h.showVars(e, node)
		}
	}
	return continueResult
}

func (h *debugHandler) matchBreakPath(path string, points []string) bool {
	for _, point := range points {
		if strings.HasSuffix(path, point) {
			return true
		}
	}
	return false
}

func (h *debugHandler) readCommand(app *kingpin.Application) (string, error) {
	fmt.Fprintf(h.out, "\n> ")

	buf := bufio.NewReader(h.in)

	line, err := buf.ReadBytes('\n')

	if err != nil {
		return "", err
	}

	line = bytes.TrimRight(line, "\n")

	args := strings.Fields(string(line))

	cmd, err := app.Parse(args)

	if err != nil {
		fmt.Fprintf(h.out, "\ninvalid command: '%v'\n", strings.Join(args, " "))
		app.Usage([]string{})
	}

	return cmd, err
}

type breakApp struct {
	app         *kingpin.Application
	cmdContinue *kingpin.CmdClause
	cmdErrors   *kingpin.CmdClause
	cmdVars     *kingpin.CmdClause
}

type failureApp struct {
	app         *kingpin.Application
	cmdContinue *kingpin.CmdClause
	cmdRetry    *kingpin.CmdClause
	cmdErrors   *kingpin.CmdClause
	cmdVars     *kingpin.CmdClause
}

func (h *debugHandler) makeBreakApp() *breakApp {
	kapp := h.makeBaseApp()
	app := &breakApp{app: kapp}
	app.cmdContinue = kapp.
		Command("continue", "continue execution").Alias("c")
	app.cmdErrors = kapp.
		Command("errors", "show errors").Alias("e")
	app.cmdVars = kapp.
		Command("vars", "show vars").Alias("v")
	return app
}

func (h *debugHandler) makeFailureApp() *failureApp {
	kapp := h.makeBaseApp()
	app := &failureApp{app: kapp}
	app.cmdContinue = kapp.
		Command("continue", "continue execution").Alias("c")
	app.cmdRetry = kapp.
		Command("retry", "retry component").Alias("r")
	app.cmdErrors = kapp.
		Command("errors", "show errors").Alias("e")
	app.cmdVars = kapp.
		Command("vars", "show vars").Alias("v")
	return app
}

func (h *debugHandler) makeBaseApp() *kingpin.Application {
	app := kingpin.New("debugger", "gestalt debugger").
		Terminate(func(n int) {}).
		UsageTemplate(usageTemplate).
		Writer(h.out)

	app.HelpFlag = app.HelpFlag.Hidden()

	return app
}

func (h *debugHandler) showErrors(e Evaluator, _ Component, curerr ...error) {

	errors := e.Errors()

	fmt.Fprintf(h.out, "%v errors\n", len(errors)+len(curerr))

	for _, err := range curerr {
		fmt.Fprintf(h.out, "[%v]\n", e.Path())
		fmt.Fprintf(h.out, "%v\n", err)
		if errd, ok := err.(ErrorWithDetail); ok {
			fmt.Fprintf(h.out, "%v\n", errd.Detail())
		}
	}

	for _, err := range errors {
		fmt.Fprintf(h.out, "%v\n", err)
		if errd, ok := err.(ErrorWithDetail); ok {
			fmt.Fprintf(h.out, "%v\n", errd.Detail())
		}
	}
}

func (h *debugHandler) showVars(e Evaluator, _ Component) {
	vars := e.Vars()
	for _, k := range vars.Keys() {
		fmt.Fprintf(h.out, "%v=%v\n", k, vars.Get(k))
	}
}

var usageTemplate = `
{{define "FormatCommand"}}\
{{if .FlagSummary}} {{.FlagSummary}}{{end}}\
{{range .Args}} {{if not .Required}}[{{end}}<{{.Name}}>{{if .Value|IsCumulative}}...{{end}}{{if not .Required}}]{{end}}{{end}}\
{{end}}\
{{define "FormatCommandList"}}\
{{range .}}\
{{if not .Hidden}}\
{{.Depth|Indent}}{{.Name}}{{if .Default}}*{{end}}{{template "FormatCommand" .}}
{{end}}\
{{template "FormatCommandList" .Commands}}\
{{end}}\
{{end}}\
{{define "FormatUsage"}}\
{{template "FormatCommand" .}}{{if .Commands}} <command> [<args> ...]{{end}}
{{if .Help}}
{{.Help|Wrap 0}}\
{{end}}\
{{end}}\
{{if .Context.SelectedCommand}}\
usage: {{.Context.SelectedCommand}}{{template "FormatUsage" .Context.SelectedCommand}}
{{if .Context.Flags}}\
Flags:
{{.Context.Flags|FlagsToTwoColumns|FormatTwoColumns}}
{{end}}\
{{if .Context.Args}}\
Args:
{{.Context.Args|ArgsToTwoColumns|FormatTwoColumns}}
{{end}}\
{{if .Context.SelectedCommand.Commands}}\
Commands:
  {{.Context.SelectedCommand}}
{{template "FormatCommandList" .Context.SelectedCommand.Commands}}
{{end}}\
{{else if .App.Commands}}\
Commands:
{{template "FormatCommandList" .App.Commands}}
{{end}}\
`
