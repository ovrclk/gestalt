package util

import (
	"fmt"
	"net"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/vars"
)

func DNSLookup(ref vars.Ref) gestalt.Component {
	ipVar := fmt.Sprintf("%s-ip", ref.Name())
	return gestalt.NewComponent("dns-lookup", func(e gestalt.Evaluator) error {

		host := ref.Expand(e.Vars())
		hosts, err := net.LookupHost(host)
		if err != nil {
			return err
		}

		if len(hosts) < 1 {
			return fmt.Errorf("dns lookup of host %v failed", host)
		}
		fmt.Println(hosts)

		e.Vars().Put(ipVar, hosts[0])

		return nil
	}).WithMeta(vars.NewMeta().Require(ref.Name()).Export(ipVar))
}
