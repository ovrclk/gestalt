package result

const (
	StateInvalid State = iota
	StateRunning
	StateComplete
	StateError
)

type State int

type Result interface {
	State() State
	Err() error
	Wait() Result
}

type result struct {
	state State
	err   error
	fn    func() Result
}

func Complete() Result {
	return &result{StateComplete, nil, nil}
}

func Error(err error) Result {
	return &result{StateError, err, nil}
}

func Running(fn func() Result) Result {
	return &result{StateRunning, nil, fn}
}

func (r *result) State() State {
	return r.state
}

func (r *result) Wait() Result {
	if r.fn != nil {
		return r.fn()
	}
	return r
}

func (r *result) Err() error {
	return r.err
}
