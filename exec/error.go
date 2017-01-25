package exec

import (
	"bytes"
	"fmt"
	"strings"
)

type Error struct {
	message string
	path    string
	args    []string

	stdout string
	stderr string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%v %v: %v", e.path, strings.Join(e.args, " "), e.message)
}

func (e *Error) Stdout() string {
	return e.stdout
}

func (e *Error) Stderr() string {
	return e.stderr
}

func (e *Error) Detail() string {
	buf := new(bytes.Buffer)
	buf.WriteString("-===[BEGIN STDOUT]===-\n")
	buf.WriteString(e.stdout)
	buf.WriteString("\n-===[END STDOUT]===-\n")
	buf.WriteString("\n-===[BEGIN STDERR]===-\n")
	buf.WriteString(e.stderr)
	buf.WriteString("\n-===[END STDERR]===-\n")
	return buf.String()
}
