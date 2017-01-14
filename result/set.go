package result

import "fmt"

type Set interface {
	IsRunning() bool
	IsError() bool
	IsComplete() bool

	Add(Result) Set
	Result() Result
	Wait() Result
}

func NewSet() *set {
	return &set{}
}

type set struct {
	children []Result
	errors   []Result
	running  []Result
}

func (s *set) Add(child Result) Set {
	switch child.State() {
	case StateError:
		s.errors = append(s.errors, child)
	case StateRunning:
		s.running = append(s.running, child)
	}
	s.children = append(s.children, child)
	return s
}

func (s *set) IsError() bool {
	return len(s.errors) > 0
}

func (s *set) IsRunning() bool {
	return len(s.running) > 0
}

func (s *set) IsComplete() bool {
	return !s.IsError() && !s.IsRunning()
}

func (s *set) Result() Result {
	return s.combine()
}

func (s *set) Wait() Result {
	return s.combine().Wait()
}

func (s *set) combine() Result {
	if len(s.running) > 0 {
		return Running(func() Result {
			sfinal := NewSet()
			for _, child := range s.children {
				sfinal.Add(child.Wait())
			}
			return sfinal.Result()
		})
	}
	if len(s.errors) > 0 {
		return Error(fmt.Errorf("%v", s.errors))
	}
	return Complete()
}
