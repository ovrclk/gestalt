package gestalt

type Traverser interface {
	Path() string
}

type Visitor interface {
	Push(Traverser, Component)
	Pop(Traverser, Component)
}

type traverser struct {
	path     *pathVisitor
	visitors []Visitor
}

func Traverse(node Component, visitors ...Visitor) {
	newTraverser(visitors...).Traverse(node)
}

func newTraverser(visitors ...Visitor) *traverser {
	path := newPathVisitor()
	return &traverser{
		path:     path,
		visitors: append([]Visitor{path}, visitors...),
	}
}

func (t *traverser) Traverse(node Component) {
	for _, v := range t.visitors {
		v.Push(t, node)
	}

	if c, ok := node.(CompositeComponent); ok {
		for _, c := range c.Children() {
			t.Traverse(c)
		}
	}

	for i := len(t.visitors) - 1; i >= 0; i-- {
		t.visitors[i].Pop(t, node)
	}
}

func (t *traverser) Path() string {
	return t.path.Current()
}
