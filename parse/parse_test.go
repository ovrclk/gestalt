package parse_test

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/ovrclk/gestalt/parse"
	"github.com/stretchr/testify/require"
)

func TestParse_Noop(t *testing.T) {

	snippet := `
name: foo
type: noop
`

	reg := parse.NewRegistry()
	reg.Register("noop",

	js, err := yaml.YAMLToJSON([]byte(snippet))
	require.NoError(t, err)

	t.Logf("%s", string(js))

	//fmt.Println(snippet)
	t.FailNow()
}
