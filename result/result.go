package result

import "fmt"

const (
	StateInvalid State = iota
	StateComplete
	StateError
)

var stateString = []string{
	"invalid",
	"complete",
	"error",
}

type State int

type Result interface {
	State() State
	IsComplete() bool
	IsError() bool
	Err() error
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

func (r *result) State() State {
	return r.state
}
func (r *result) IsComplete() bool {
	return r.state == StateComplete
}
func (r *result) IsError() bool {
	return r.state == StateError
}

func (r *result) Err() error {
	return r.err
}

func (r *result) String() string {
	if r.state == StateError {
		return fmt.Sprintf("[%v: %v]", stateString[r.state], r.err)
	}
	return fmt.Sprintf("[%v]", stateString[r.state])
}
