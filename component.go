package gestalt

type Component interface {
	Name() string
	Children() []Component
	AddChild(Component) Component
	Runable(Runner) Runable
}

type component struct {
	name     string
	children []Component
}

func NewComponent(name string) *component {
	return &component{name: name}
}

func (c *component) Name() string {
	return c.name
}

func (c *component) Children() []SuiteComponent {
	return c.children
}

func (c *component) AddChild(child Component) Component {
	c.children = append(c.children, child)
	return child
}
