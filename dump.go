package gestalt

import "fmt"

func Dump(c Component) {
	Traverse(c, newDumper())
}

type dumper struct {
}

func newDumper() *dumper {
	return &dumper{}
}

func (d *dumper) Push(t Traverser, c Component) {
	fmt.Printf("%v\n", t.Path())
}

func (d *dumper) Pop(_ Traverser, _ Component) {
}
