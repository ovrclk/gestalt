package gestalt

import (
	"fmt"
	"os"
	"time"

	"github.com/ovrclk/gestalt/vars"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func Run(c Component) {
	RunWith(c, os.Args[1:])
}

func RunWith(c Component, args []string) {
	NewRunner().
		WithArgs(args).
		WithComponent(c).
		Run()
}

type Runner interface {
	WithArgs([]string) Runner
	WithComponent(Component) Runner
	WithTerminate(func(status int)) Runner
	Run()
}

type runner struct {
	args []string
	cmp  Component

	terminate func(status int)
}

func NewRunner() Runner {
	return &runner{terminate: os.Exit}
}

func (r *runner) WithArgs(args []string) Runner {
	r.args = args
	return r
}

func (r *runner) WithComponent(cmp Component) Runner {
	r.cmp = cmp
	return r
}

func (r *runner) WithTerminate(terminate func(int)) Runner {
	r.terminate = terminate
	return r
}

func (r *runner) Run() {

	opts := newOptions(r)

	if r.cmp == nil {
		opts.app.Fatalf("no component given")
		return
	}

	switch kingpin.MustParse(opts.app.Parse(r.args)) {
	case opts.cmdShow.FullCommand():
		r.doShow(opts)
	case opts.cmdEval.FullCommand():
		r.doEval(opts)
	case opts.cmdValidate.FullCommand():
		r.doValidate(opts)
	}
}

type options struct {
	app *kingpin.Application

	logLevel *string
	logFile  **os.File

	vars *map[string]string

	cmdShow *kingpin.CmdClause

	cmdEval *kingpin.CmdClause
	trace   *bool

	breakpoints *[]string
	failpoints  *[]string

	cmdValidate *kingpin.CmdClause
}

func (opts *options) getVars() vars.Vars {
	v := vars.NewVars()
	if opts.vars != nil {
		v = v.Merge(vars.FromMap(*opts.vars))
	}
	return v
}

func newOptions(r *runner) *options {
	opts := &options{}

	opts.app = kingpin.
		New("gestalt", "Run gestalt components").
		Terminate(r.terminate)

	opts.logLevel = opts.app.
		Flag("log-level", "Log level").Short('l').
		Default("panic").
		Enum("panic", "fatal", "error", "warn", "info", "debug")

	opts.logFile = opts.app.
		Flag("log-file", "Write log to file").
		Default("/dev/stdout").
		OpenFile(os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)

	opts.vars = opts.app.
		Flag("set", "set variables").Short('s').StringMap()

	opts.cmdEval = opts.app.
		Command("eval", "run components").Default()

	opts.trace = opts.cmdEval.
		Flag("trace", "Trace execution").
		Bool()

	opts.breakpoints = opts.cmdEval.
		Flag("breakpoint", "add breakpoint").
		Short('B').
		Strings()

	opts.failpoints = opts.cmdEval.
		Flag("failpoint", "breakpoint after failure").
		Short('b').
		Strings()

	opts.cmdShow = opts.app.
		Command("show", "display component tree")

	opts.cmdValidate = opts.app.
		Command("validate", "validate vars")

	return opts
}

func (r *runner) doEval(opts *options) {

	lb := newLogBuilder().
		WithLevel(*opts.logLevel).
		WithLogOut(*opts.logFile)

	profiler := newProfileVisitor()

	visitors := []Visitor{profiler}

	if *opts.trace {
		visitors = append(visitors, newTraceVisitor(os.Stdout))
	}

	e := NewEvaluatorWithLogger(lb.Logger(), visitors...)

	if opts.breakpoints != nil || opts.failpoints != nil {
		handler := newDebugHandler(os.Stdin, os.Stdout)

		if opts.breakpoints != nil {
			for _, point := range *opts.breakpoints {
				handler.AddBreakpoint(point)
			}
		}
		if opts.failpoints != nil {
			for _, point := range *opts.failpoints {
				handler.AddFailpoint(point)
			}
		}

		e.handler = handler
	}

	e.Vars().Merge(opts.getVars())

	if err := r.showUnresolvedVars(opts, e.Vars()); err != nil {
		opts.app.FatalIfError(err, "")
	}

	e.Evaluate(r.cmp)
	e.Wait()

	// show profile info
	fmt.Printf("\nprofile info:\n\n")
	for _, p := range profiler.profiles {
		fmt.Printf("%-5v%-10v%v\n", p.count, p.avg/time.Millisecond, p.path)
	}

	if !e.HasError() {
		fmt.Printf("\nall tests passed\n")
		return
	}

	fprintErr(os.Stderr, "\n\nEvaluation of %v failed:\n\n", r.cmp.Name())

	for _, err := range e.Errors() {
		if err, ok := err.(Error); ok {

			fmt.Fprintf(os.Stderr, "%v\n", err)
			fmt.Fprintf(os.Stderr, "%v\n", err.Detail())
		} else {
			fmt.Fprintf(os.Stderr, "Unknown Error:\n%v\n", err)
		}
	}
	opts.app.Fatalf("eval failed")
}

func (r *runner) doShow(opts *options) {
	Dump(r.cmp)
}

func (r *runner) doValidate(opts *options) {
	err := r.showUnresolvedVars(opts, opts.getVars())
	opts.app.FatalIfError(err, "")
}

func (r *runner) showUnresolvedVars(opts *options, vars vars.Vars) error {

	unresolved := ValidateWith(r.cmp, vars)

	if len(unresolved) == 0 {
		return nil
	}

	for _, x := range unresolved {
		opts.app.Errorf("Missing variables:\n")
		opts.app.Errorf("%v.%v", x.Path, x.Name)
	}

	return fmt.Errorf("missing variables")
}
