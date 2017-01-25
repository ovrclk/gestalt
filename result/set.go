package result

import "fmt"

type Set interface {
	IsError() bool
	IsComplete() bool

	Add(Result) Set
	Result() Result
}

func NewSet() *set {
	return &set{}
}

type set struct {
	children []Result
	errors   []Result
}

func (s *set) Add(child Result) Set {
	switch child.State() {
	case StateError:
		s.errors = append(s.errors, child)
	}
	s.children = append(s.children, child)
	return s
}

func (s *set) IsError() bool {
	return len(s.errors) > 0
}

func (s *set) IsComplete() bool {
	return !s.IsError()
}

func (s *set) Result() Result {
	return s.combine()
}

func (s *set) combine() Result {
	if len(s.errors) > 0 {
		return Error(fmt.Errorf("%v", s.errors))
	}
	return Complete()
}
