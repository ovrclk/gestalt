package util

import (
	"fmt"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/exec"
	"github.com/ovrclk/gestalt/vars"
)

func DNSLookup(ref vars.Ref) gestalt.Component {
	ipVar := fmt.Sprintf("%s-ip", ref.Name())
	return exec.EXEC("dns-lookup", "dig", ref.Var(), "+short").
		FN(exec.Capture(ipVar)).
		WithMeta(vars.NewMeta().Require(ref.Name()).Export(ipVar))
}
