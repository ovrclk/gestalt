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
	return h.runDebugger(e, node, h.makeBreakApp)
}

func (h *debugHandler) runFailureConsole(e Evaluator, node Component, err error) commandResult {
	return h.runDebugger(e, node, h.makeFailureApp, err)
}

func (h *debugHandler) runDebugger(
	e Evaluator,
	node Component,
	appBuilder func() *debugApp, errors ...error) commandResult {

	for i := 0; ; i++ {
		app := appBuilder()

		if i == 0 {
			app.app.Usage([]string{})
		}

		fmt.Fprintf(h.out, "\n%v: %v errors\n", e.Path(), len(e.Errors())+len(errors))

		cmd, err := h.readCommand(app.app)
		if err == io.EOF {
			return continueResult
		}

		check := func(clause *kingpin.CmdClause, cmd string) bool {
			return clause != nil && clause.FullCommand() == cmd
		}

		switch {
		case check(app.cmdContinue, cmd):
			fmt.Fprintf(h.out, "continuing...\n")
			return continueResult
		case check(app.cmdRetry, cmd):
			fmt.Fprintf(h.out, "retrying...\n")
			return retryResult
		case check(app.cmdErrors, cmd):
			h.showErrors(e, node, errors...)
		case check(app.cmdVars, cmd):
			h.showVars(e, node)

		// breakpoints
		case check(app.cmdBPList, cmd):
			h.showPoints(h.breakpoints)
		case check(app.cmdBPAdd, cmd):
			h.breakpoints = append(h.breakpoints, *app.cmdBPAddEntries...)
			h.showPoints(h.breakpoints)
		case check(app.cmdBPDel, cmd):
			h.breakpoints = h.delPoints(h.breakpoints, *app.cmdBPDelEntries)
			h.showPoints(h.breakpoints)

		// failpoints
		case check(app.cmdFPList, cmd):
			h.showPoints(h.failpoints)
		case check(app.cmdFPAdd, cmd):
			h.failpoints = append(h.failpoints, *app.cmdFPAddEntries...)
			h.showPoints(h.failpoints)
		case check(app.cmdFPDel, cmd):
			h.failpoints = h.delPoints(h.failpoints, *app.cmdFPDelEntries)
			h.showPoints(h.failpoints)
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

	if len(line) == 0 {
		return "", nil
	}

	args := strings.Fields(string(line))

	cmd, err := app.Parse(args)

	if err != nil {
		fmt.Fprintf(h.out, "\ninvalid command: '%v'\n", strings.Join(args, " "))
		app.Usage([]string{})
	}

	return cmd, err
}

type debugApp struct {
	app         *kingpin.Application
	cmdContinue *kingpin.CmdClause
	cmdRetry    *kingpin.CmdClause
	cmdErrors   *kingpin.CmdClause
	cmdVars     *kingpin.CmdClause

	cmdBP           *kingpin.CmdClause
	cmdBPList       *kingpin.CmdClause
	cmdBPAdd        *kingpin.CmdClause
	cmdBPAddEntries *[]string
	cmdBPDel        *kingpin.CmdClause
	cmdBPDelEntries *[]uint

	cmdFP           *kingpin.CmdClause
	cmdFPList       *kingpin.CmdClause
	cmdFPAdd        *kingpin.CmdClause
	cmdFPAddEntries *[]string
	cmdFPDel        *kingpin.CmdClause
	cmdFPDelEntries *[]uint
}

func (h *debugHandler) makeBreakApp() *debugApp {
	kapp := h.makeBaseApp()
	app := &debugApp{app: kapp}
	app.cmdContinue = kapp.
		Command("continue", "continue execution").Alias("c")
	app.cmdErrors = kapp.
		Command("errors", "show errors").Alias("e")
	app.cmdVars = kapp.
		Command("vars", "show current vars").Alias("v")

	// breakpoint commands
	app.cmdBP = kapp.
		Command("breakpoint", "manipulate breakpoints").Alias("bp")
	app.cmdBPList = app.cmdBP.
		Command("list", "show current breakpoints").Alias("l").Default()
	app.cmdBPAdd = app.cmdBP.
		Command("add", "add breakpoints").Alias("a")
	app.cmdBPAddEntries = app.cmdBPAdd.
		Arg("pattern", "breakpoint pattern to add").
		Strings()
	app.cmdBPDel = app.cmdBP.
		Command("del", "delete breakpoints").Alias("d")
	app.cmdBPDelEntries = app.cmdBPDel.
		Arg("index", "breakpoint numbers to delete").
		Uints()

	// failpoint commands
	app.cmdFP = kapp.
		Command("failpoint", "manipulate failpoints").Alias("fp")
	app.cmdFPList = app.cmdFP.
		Command("list", "show current failpoints").Alias("l").Default()
	app.cmdFPAdd = app.cmdFP.
		Command("add", "add failpoint").Alias("a")
	app.cmdFPAddEntries = app.cmdFPAdd.
		Arg("pattern", "failpoint pattern to add").
		Strings()
	app.cmdFPDel = app.cmdFP.
		Command("del", "delete failpoints").Alias("d")
	app.cmdFPDelEntries = app.cmdFPDel.
		Arg("index", "failpoint numbers to delete").
		Uints()

	return app
}

func (h *debugHandler) makeFailureApp() *debugApp {
	app := h.makeBreakApp()
	app.cmdRetry = app.app.
		Command("retry", "retry component").Alias("r")
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

func (h *debugHandler) showPoints(points []string) {
	for i, point := range points {
		fmt.Fprintf(h.out, "[%v]\t%v\n", i, point)
	}
}

func (h *debugHandler) delPoints(current []string, indexes []uint) []string {
	points := make([]string, 0)

	for i, point := range current {
		keep := true
		for _, j := range indexes {
			if j == uint(i) {
				keep = false
				break
			}
		}
		if keep {
			points = append(points, point)
		}
	}
	return points
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
