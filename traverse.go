package gestalt

func Traverse(node Component, fn func(Component)) {
	fn(node)
}
