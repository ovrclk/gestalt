package gestalt

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
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
	app   *kingpin.Application
	debug *bool

	vars *map[string]string

	cmdShow     *kingpin.CmdClause
	cmdEval     *kingpin.CmdClause
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

	opts.app = kingpin.New("gestalt", "Run gestalt components").Terminate(r.terminate)
	opts.debug = opts.app.Flag("debug", "Display debug logging").Bool()

	opts.vars = opts.app.Flag("set", "set variables").Short('s').StringMap()

	opts.cmdEval = opts.app.Command("eval", "run components").Default()
	opts.cmdShow = opts.app.Command("show", "display component tree")
	opts.cmdValidate = opts.app.Command("validate", "validate vars")

	return opts
}

func (r *runner) doEval(opts *options) {

	logger := logrus.New()
	if *opts.debug {
		logger.Level = logrus.DebugLevel
	}

	e := NewEvaluatorWithLogger(logger)

	e.Vars().Merge(opts.getVars())

	if err := r.showUnresolvedVars(opts, e.Vars()); err != nil {
		opts.app.FatalIfError(err, "")
	}

	err := e.Evaluate(r.cmp).Wait().Err()

	opts.app.FatalIfError(err, "error evaluating components")
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
