package util

import (
	"github.com/ovrclk/gestalt/exec"
)

func HTTPGet(host string) exec.Cmd {
	return exec.EXEC("http-get", "curl", "-s", host)
}
