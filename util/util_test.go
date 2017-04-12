package util_test

import (
	"testing"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/exec"
	"github.com/ovrclk/gestalt/util"
	"github.com/ovrclk/gestalt/vars"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDNSLookup(t *testing.T) {
	ref := vars.NewRef("host")

	e := gestalt.NewEvaluator()
	e.Vars().Put(ref.Name(), "google-public-dns-a.google.com")

	require.NoError(t, e.Evaluate(util.DNSLookup(ref)))
	require.Empty(t, e.Errors())

	assert.True(t, e.Vars().Has("host-ip"))
	assert.Equal(t, "8.8.8.8", e.Vars().Get("host-ip"))
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
