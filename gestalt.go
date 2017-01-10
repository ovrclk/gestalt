package gestalt

import "github.com/Sirupsen/logrus"

// Component

// Runable
//  - Start
//  - Stop

type SuiteComponent interface {
	Name() string

	Initialize() error

	Run(RunContext) Result

	Cleanup() error

	AddChild(SuiteComponent) SuiteComponent
}

type RunState int

const (
	RunStateStopped RunState = iota
	RunStateRunning
	RunStateComplete
	RunStateError
)

type Result interface {
	State() RunState
}

type RunContext interface {
}

type Component struct {
	Name     string
	children []SuiteComponent
}

type Suite struct {
	Component
	logger logrus.FieldLogger
}

func (c *Component) ComponentName() string {
	return c.Name
}

func (c *Component) Children() []SuiteComponent {
	return c.children
}

func (c *Component) AddChild(child SuiteComponent) SuiteComponent {
	c.children = append(c.children, child)
	return child
}

func (c *Component) Cleanup() error {
	return nil
}

type CommandComponent struct {
	Component
	Path string
	Args []string
	Env  []string
}

type CommandResult struct {
}

func NewCommandComponent(path string, args []string) *CommandComponent {
	return &CommandComponent{
		Component: Component{Name: path},
		Path:      path,
		Args:      args,
	}
}

func (c *CommandComponent) Initialize() error {
	return nil
}

func (c *CommandComponent) Run(rctx RunContext) Result {
	return nil
}

func NewSuite(name string) *Suite {
	suite := &Suite{
		Component: Component{Name: name},
		logger: logrus.WithFields(logrus.Fields{
			"suite": name,
		}),
	}
	return suite
}

func (s *Suite) Run() {
	for _, child := range s.Children() {
		s.runComponent(child)
	}
}

func (s *Suite) runComponent(c SuiteComponent) {
	s.logger.
		WithField("component", c.ComponentName()).
		WithField("event", "start").
		Infof("starting %v", c.ComponentName())
}
