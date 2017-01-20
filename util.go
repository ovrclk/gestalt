package gestalt

import "fmt"

func pushPath(path string, c Component) string {
	if c.IsPassThrough() {
		return path
	} else {
		return fmt.Sprintf("%v/%v", path, c.Name())
	}
}
