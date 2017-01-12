package gestalt

type Suite struct {
	component
}

func NewSuite(name string) *Suite {
	return &Suite{component{name: name}}
}

func (s *Suite) Build(bctx BuildCtx) *Runable {
	return NullRunable()
}
