package vars

import "fmt"

type Ref string

func NewRef(name string) Ref {
	return Ref(name)
}

func (ref Ref) Name() string {
	return string(ref)
}

func (ref Ref) Var() string {
	return fmt.Sprintf("{{%s}}", ref.Name())
}

func (ref Ref) String() string {
	return string(ref)
}
