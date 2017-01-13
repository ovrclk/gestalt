package gestalt

const (
	RunStateStopped RunState = iota
	RunStateRunning
	RunStateComplete
	RunStateError
)

type RunState int

type Result interface {
	State() RunState
	Err() error
	Values() ResultValues
	Wait()
}

type ResultValues map[string]interface{}

type result struct {
	state  RunState
	values ResultValues
	err    error
	wait   func()
}

func NewResult(state RunState, values ResultValues, err error) *result {
	return &result{state, values, err, nil}
}

func ResultSuccess() *result {
	return NewResult(RunStateComplete, nil, nil)
}

func ResultError(err error) *result {
	return NewResult(RunStateError, nil, err)
}

func ResultRunning(f func()) *result {
	return &result{RunStateRunning, nil, nil, f}
}

func (r *result) State() RunState {
	return r.state
}

func (r *result) Err() error {
	return r.err
}

func (r *result) Values() ResultValues {
	return r.values
}

func (r *result) Wait() {
	if r.state == RunStateRunning && r.wait != nil {
		r.wait()
	}
}