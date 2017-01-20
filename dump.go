package gestalt

import "fmt"

func Dump(c Component) {
	Walk(c, newDumper())
}

type dumper struct {
	depth int
}

func newDumper() *dumper {
	return &dumper{depth: -1}
}

func (d *dumper) Push(c Component) {
	d.depth += 1
	d.display(c)
}

func (d *dumper) Pop(c Component) {
	d.depth -= 1
}

func (d *dumper) display(c Component) {
	fmt.Printf("%*s", d.depth*2+1, "")
	fmt.Printf("- %s", c.Name())
	fmt.Printf("\n")
}
