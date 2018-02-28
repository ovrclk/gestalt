package util_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/exec"
	"github.com/ovrclk/gestalt/util"
	"github.com/ovrclk/gestalt/vars"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDNSLookup(t *testing.T) {
	tests := map[string]string{
		// A
		"google-public-dns-a.google.com": "8.8.8.8",

		// CNAME
		//"motogp.reddit.com": "151.101.1.140",
	}
	ref := vars.NewRef("host")

	for host, ip := range tests {
		e := gestalt.NewEvaluator()
		e.Vars().Put(ref.Name(), host)

		if !assert.NoError(t, e.Evaluate(util.DNSLookup(ref)), host) {
			continue
		}

		if !assert.Empty(t, e.Errors(), host) {
			continue
		}
		assert.True(t, e.Vars().Has("host-ip"))
		assert.Equal(t, ip, e.Vars().Get("host-ip"))
	}
}

func TestDNSLookupFailure(t *testing.T) {
	// bogus host
	host := fmt.Sprintf("%v.%v", time.Now().Unix(), time.Now().Unix())
	ref := vars.NewRef("host")
	e := gestalt.NewEvaluator()
	e.Vars().Put(ref.Name(), host)

	assert.Error(t, e.Evaluate(util.DNSLookup(ref)))
	assert.NotEmpty(t, e.Errors())
}

func TestHTTPGet(t *testing.T) {
	host := "beta.release.core-os.net"
	url := "https://{{host}}/amd64-usr/1353.4.0/version.txt"

	e := gestalt.NewEvaluator()
	e.Vars().Put("host", host)

	cmp := util.HTTPGet(url).
		FN(exec.ParseKV("=", "key", "value").
			GrepField("key", "COREOS_BUILD").
			EnsureCount(1).
			CaptureAll()).
		WithMeta(vars.NewMeta().Export("key", "value"))

	require.NoError(t, e.Evaluate(cmp))
	require.Empty(t, e.Errors())

	assert.True(t, e.Vars().Has("key"))
	assert.True(t, e.Vars().Has("value"))

	assert.Equal(t, "COREOS_BUILD", e.Vars().Get("key"))
	assert.Equal(t, "1353", e.Vars().Get("value"))
}
