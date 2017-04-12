package util_test

import (
	"testing"

	"github.com/ovrclk/gestalt"
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

	assert.True(t, e.Vars().Has("host-ip"))
	assert.Equal(t, "8.8.8.8", e.Vars().Get("host-ip"))
}
