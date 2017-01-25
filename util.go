package gestalt

import "fmt"

func pushPath(path string, c Component) string {
	if c.IsPassThrough() {
		return path
	} else {
		return fmt.Sprintf("%v/%v", path, c.Name())
	}
}

func adornPath(path string, node Component) string {
	if node.IsPassThrough() {
		path += "/" + node.Name()
	}
	return path
}
