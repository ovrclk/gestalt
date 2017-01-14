package gestalt

import "fmt"

type Dumper interface {
	Dump(Component)
}

type dumper struct {
	depth int
}

func Dump(c Component) {
	NewDumper().Dump(c)
}

func NewDumper() Dumper {
	return &dumper{depth: -1}
}

func (d *dumper) Dump(c Component) {
	d.cloneFor(c).dump(c)
}

func (d *dumper) cloneFor(_ Component) *dumper {
	return &dumper{depth: d.depth + 1}
}

func (d *dumper) dump(c Component) {
	fmt.Printf("%*s", d.depth*2+1, "")
	fmt.Printf("- %s", c.Name())
	fmt.Printf("\n")

	if parent, ok := c.(CompositeComponent); ok {
		for _, child := range parent.Children() {
			d.Dump(child)
		}
	}
}
