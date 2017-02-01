package gestalt

import (
	"io"

	"github.com/fatih/color"
)

func fprintErr(w io.Writer, fmt string, args ...interface{}) {
	clr := color.New(color.FgHiRed)
	clr.Fprintf(w, fmt, args...)
}
