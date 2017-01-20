package gestalt

type Walker interface {
	Push(Component)
	Pop(Component)
}

func Walk(c Component, walker Walker) {
	walker.Push(c)
	if c, ok := c.(CompositeComponent); ok {
		for _, c := range c.Children() {
			Walk(c, walker)
		}
	}
	walker.Pop(c)
}
