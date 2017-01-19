package vars_test

import (
	"reflect"
	"testing"

	"github.com/ovrclk/gestalt/vars"
)

func TestImportTo(t *testing.T) {
	parent := vars.NewVars()
	parent.Put("a", "foo")

	child := vars.NewVars()

	m := vars.NewMeta().Require("a")

	vars.ImportTo(m, parent, child)

	if val := child.Get("a"); val != "foo" {
		t.Errorf("key not imported into child")
	}
}

func TestExportTo(t *testing.T) {
	child := vars.NewVars()
	child.Put("a", "foo")

	parent := vars.NewVars()

	m := vars.NewMeta().Export("a")

	vars.ExportTo(m, child, parent)

	if val := parent.Get("a"); val != "foo" {
		t.Errorf("key not exported to parent")
	}
}

func TestMerge(t *testing.T) {
	m1 := vars.NewMeta()
	m1.Require("a")

	m2 := vars.NewMeta()
	m2.Export("b")

	m3 := m1.Merge(m2)

	if !reflect.DeepEqual(m3.Exports(), []string{"b"}) {
		t.Errorf("merge result does not have exports (%v)", m3.Exports())
	}

	if !reflect.DeepEqual(m3.Requires(), []string{"a"}) {
		t.Errorf("merge result does not have exports (%v)", m3.Requires())
	}
}
