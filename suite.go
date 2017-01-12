package gestalt

type Suite struct {
	component
}

func NewSuite(name string) *Suite {
	return &Suite{component{name: name}}
}

func (s *Suite) Build(bctx BuildCtx) *Runable {
	return nil
}

func (s *Suite) IsTerminal() bool {
	return true
}

type Group struct {
	component
}

func NewGroup(name string) *Group {
	return &Group{component{name: name}}
}

func (g *Group) IsTerminal() bool {
	return false
}

func (g *Group) Build(bctx BuildCtx) *Runable {
	return nil
}
