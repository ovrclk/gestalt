package exec

import "strings"

func EXEC(name, path string, args ...string) *Cmd {
	return NewCmd(name, path, args)
}

func SH(name, cmd string, args ...string) *Cmd {
	return NewCmd(
		name,
		"/bin/sh",
		[]string{
			"-c",
			strings.Join(append([]string{cmd}, args...), " "),
		})
}
