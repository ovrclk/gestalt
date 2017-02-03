package exec_test

import (
	"bufio"
	"bytes"
	"context"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/ovrclk/gestalt"
	"github.com/ovrclk/gestalt/exec"
	"github.com/ovrclk/gestalt/vars"
)

func TestParse(t *testing.T) {
	e := newEvaluator(t)
	p := exec.ParseColumns("a", "b")
	b := bytes.NewBufferString("foo bar")
	err := p.CaptureAll()(bufio.NewReader(b), e)

	if err != nil {
		t.Error(err)
	}

	if x := e.vars.Get("a"); x != "foo" {
		t.Fatalf("invalid export: field a of %v != %v", e.vars, "foo")
	}
}

func TestGrepField(t *testing.T) {
	e := newEvaluator(t)
	b := bytes.NewBufferString("foo bar\nbar baz\nxyz abc")

	p := exec.ParseColumns("a", "b")
	p.GrepField("a", "bar")

	err := p.CaptureAll()(bufio.NewReader(b), e)

	if err != nil {
		t.Error(err)
	}

	if x := e.vars.Get("b"); x != "baz" {
		t.Fatalf("invalid export: field b of %v != %v", e.vars, "baz")
	}
}

func TestEnsureCount(t *testing.T) {
	{
		e := newEvaluator(t)
		b := bytes.NewBufferString("foo bar\nbar baz\nxyz abc")

		p := exec.ParseColumns("a", "b")
		p.GrepField("a", "bar")
		p.EnsureCount(1)

		err := p.CaptureAll()(bufio.NewReader(b), e)

		if err != nil {
			t.Error(err)
		}
	}
	{
		e := newEvaluator(t)
		b := bytes.NewBufferString("foo bar\nbar baz\nxyz abc")

		p := exec.ParseColumns("a", "b")
		p.EnsureCount(1)

		err := p.CaptureAll()(bufio.NewReader(b), e)

		if err == nil {
			t.Fatal("Expected error but non received")
		}
	}
}

func TestExpand(t *testing.T) {
	e := newEvaluator(t)

	e.vars.Put("some-value", "baz")

	b := bytes.NewBufferString("foo bar\nbar baz\nxyz abc")

	p := exec.ParseColumns("a", "b")
	p.GrepField("b", "{{some-value}}")
	p.EnsureCount(1)

	err := p.CaptureAll()(bufio.NewReader(b), e)

	if err != nil {
		t.Error(err)
	}

	if x := e.vars.Get("b"); x != "baz" {
		t.Fatalf("invalid export: field b of %v != %v", e.vars, "baz")
	}

}

func newEvaluator(t *testing.T) *fakeEvaluator {
	return &fakeEvaluator{t, vars.NewVars()}
}

type fakeEvaluator struct {
	t    *testing.T
	vars vars.Vars
}

func (e *fakeEvaluator) Log() logrus.FieldLogger {
	e.t.Fatal("Log() called")
	return nil
}

func (e *fakeEvaluator) Evaluate(gestalt.Component) error {
	e.t.Fatal("Evaluate() called")
	return nil
}

func (e *fakeEvaluator) Fork(gestalt.Component) error {
	e.t.Fatal("Fork() called")
	return nil
}
func (e *fakeEvaluator) Emit(k string, v string) {
	e.vars.Put(k, v)
}
func (e *fakeEvaluator) Vars() vars.Vars {
	return e.vars
}
func (e *fakeEvaluator) Message(string, ...interface{}) {
	e.t.Fatal("Message() called")
}
func (e *fakeEvaluator) Context() context.Context {
	e.t.Fatal("Context() called")
	return nil
}
func (e *fakeEvaluator) Stop() {
	e.t.Fatal("Stop() called")
}

func (e *fakeEvaluator) ClearError() {
	e.t.Fatal("ClearError() called")
}

func (e *fakeEvaluator) HasError() bool {
	e.t.Fatal("HasError() called")
	return false
}

func (e *fakeEvaluator) Errors() []error {
	e.t.Fatal("Errors() called")
	return nil
}

func (e *fakeEvaluator) Wait() {
	e.t.Fatal("Wait() called")
}

func (e *fakeEvaluator) Path() string {
	e.t.Fatal("Wait() called")
	return ""
}

func (e *fakeEvaluator) Root() gestalt.Component {
	e.t.Fatal("Root() called")
	return nil
}
