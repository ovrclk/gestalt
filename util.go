package gestalt

import (
	"fmt"
	"io"
	"time"

	"github.com/fatih/color"
)

func fprintErr(w io.Writer, fmt string, args ...interface{}) {
	clr := color.New(color.FgHiRed)
	clr.Fprintf(w, fmt, args...)
}

func fmtDuration(d time.Duration) string {
	mins := d / (time.Second * 60)

	d = d - mins*(time.Second*60)

	secs := d / time.Second

	d = d - secs*time.Second

	d = d / (time.Millisecond * 10)

	return fmt.Sprintf("%d:%d.%.2d", mins, secs, d)
}
