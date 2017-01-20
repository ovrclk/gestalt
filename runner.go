package gestalt

import (
	"fmt"
	"os"

	"github.com/ovrclk/gestalt/vars"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func Run(c Component) error {
	return RunWith(c, os.Args[1:])
}

func RunWith(c Component, args []string) error {
	return NewRunner().
		WithArgs(args).
		WithComponent(c).
		Run()
}

type Runner interface {
	WithArgs([]string) Runner
	WithComponent(Component) Runner
	Run() error
}

type runner struct {
	args []string
	cmp  Component
}

func NewRunner() Runner {
	return &runner{}
}

func (r *runner) WithArgs(args []string) Runner {
	r.args = args
	return r
}

func (r *runner) WithComponent(cmp Component) Runner {
	r.cmp = cmp
	return r
}

func (r *runner) Run() error {
	if r.cmp == nil {
		return fmt.Errorf("No component found")
	}

	opts := newOptions()

	switch kingpin.MustParse(opts.app.Parse(r.args)) {
	case opts.cmdShow.FullCommand():
		return r.doShow(opts)
	case opts.cmdEval.FullCommand():
		return r.doEvaluate(opts)
	}

	return fmt.Errorf("unknown command")
}

type options struct {
	app   *kingpin.Application
	debug *bool

	vars *map[string]string

	cmdShow *kingpin.CmdClause
	cmdEval *kingpin.CmdClause
}

func newOptions() *options {
	opts := &options{}

	opts.app = kingpin.New("gestalt", "Run gestalt components")
	opts.debug = opts.app.Flag("debug", "Display debug logging").Bool()

	opts.vars = opts.app.Flag("set", "set variables").Short('s').StringMap()

	opts.cmdShow = opts.app.Command("show", "display component tree")
	opts.cmdEval = opts.app.Command("eval", "run components").Default()

	return opts
}

func (r *runner) doEvaluate(opts *options) error {
	e := NewEvaluator()

	if opts.vars != nil {
		e.Vars().Merge(vars.FromMap(*opts.vars))
	}

	return e.Evaluate(r.cmp).Wait().Err()
}

func (r *runner) doShow(opts *options) error {
	Dump(r.cmp)
	return nil
}
