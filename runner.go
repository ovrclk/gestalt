package gestalt

import "context"

type Runner struct {
	ctx context.Context
}

func NewRunner() *Runner {
}

func (r *Runner) RunComponent(c Component) {
}

type Runable struct {
	retries int
}
