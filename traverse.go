package gestalt

import "fmt"

func traverse(node Component, fn func(Component)) {
	fn(node)
	switch c := node.(type) {
	case CompositeComponent:
		for _, child := range c.Children() {
			traverse(child, fn)
		}
	case WrapComponent:
		traverse(c.Child(), fn)
	}
}

func Dump(root Component) {
	dump(root, 0, func(node Component, depth int) {
		fmt.Printf("%*s", depth*2+1, "")
		fmt.Printf("- %s", node.Name())
		fmt.Printf("\n")
	})
}

func dump(node Component, depth int, fn func(Component, int)) {
	fn(node, depth)
	switch c := node.(type) {
	case CompositeComponent:
		for _, child := range c.Children() {
			dump(child, depth+1, fn)
		}
	case WrapComponent:
		dump(c.Child(), depth+1, fn)
	}
}
