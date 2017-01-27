package gestalt

import (
	"github.com/deckarep/golang-set"
	"github.com/ovrclk/gestalt/vars"
)

func Validate(c Component) []Unresolved {
	return ValidateWith(c, vars.NewVars())
}

func ValidateWith(c Component, vars vars.Vars) []Unresolved {
	v := NewValidator()
	for _, k := range vars.Keys() {
		v.top.resolved.Add(k)
	}
	Traverse(c, v)
	return v.unresolved
}

type Unresolved struct {
	Path string
	Name string
}

type validator struct {
	top        *state
	stack      []*state
	unresolved []Unresolved
}

type state struct {
	resolved mapset.Set
}

func NewValidator() *validator {
	top := &state{mapset.NewSet()}
	return &validator{
		stack: []*state{},
		top:   top,
	}
}

func (v *validator) Push(t Traverser, c Component) {

	newtop := &state{v.top.resolved.Clone()}

	for k, _ := range c.Meta().Defaults() {
		newtop.resolved.Add(k)
	}

	for _, k := range c.Meta().Requires() {
		if !v.top.resolved.Contains(k) && !newtop.resolved.Contains(k) {
			v.unresolved = append(v.unresolved, Unresolved{t.Path(), k})
		}
		newtop.resolved.Add(k)
	}

	v.stack = append(v.stack, v.top)
	v.top = newtop
}

func (v *validator) Pop(t Traverser, c Component) {
	last := len(v.stack) - 1
	v.top = v.stack[last]
	v.stack = v.stack[0:last]

	for _, k := range c.Meta().Exports() {
		v.top.resolved.Add(k)
	}
}
