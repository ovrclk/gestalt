package test

import (
	"reflect"
	"testing"

	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/vars"
)

func TestMeta(t *testing.T) {
	exports := []string{"a"}

	c := gestalt.NewComponent("foo", nil).
		WithMeta(vars.NewMeta().Export(exports...))

	if x := c.Meta().Exports(); !reflect.DeepEqual(x, exports) {
		t.Errorf("meta export missing keys %v != %v", x, exports)
	}
}
