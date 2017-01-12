package gestalt

import (
	"strings"
	"time"
)

var rootBuilder = &builder{}

type Builder interface {
	Group(string) Builder
	Suite(string) Builder
	Run(Builder) Builder
	Build() Component
}

type builder struct {
	children  []Builder
	component Component
}

type gBuilder struct {
	builder
}

type sBuilder struct {
	builder
}

type cBuilder struct {
	builder
}

func RootBuilder() Builder {
	return rootBuilder
}

func BuildGroup(name string) Builder {
	return rootBuilder.Group(name)
}

func BuildSuite(name string) Builder {
	return rootBuilder.Suite(name)
}

func (b *builder) Build() Component {
	return nil
}

func (b *builder) Group(name string) Builder {
	return &gBuilder{builder{component: &Group{component{name: name}}}}
}

func (b *builder) Suite(name string) Builder {
	return &gBuilder{builder{component: &Suite{component{name: name}}}}
}

func (b *builder) Component(c Component) Builder {
	return &cBuilder{builder{component: c}}
}

func (b *builder) Run(child Builder) Builder {
	b.children = append(b.children, child)
	return b
}

func (b *sBuilder) Build() Component {
	c := b.component
	for _, child := range b.children {
		c.AddChild(child.Build())
	}
	return c
}

func (b *gBuilder) Build() Component {
	c := b.component
	for _, child := range b.children {
		c.AddChild(child.Build())
	}
	return c
}

func (b *cBuilder) Build() Component {
	c := b.component
	for _, child := range b.children {
		c.AddChild(child.Build())
	}
	return c
}

func (b *builder) SH(name string, cmd string, args ...string) Builder {
	return b.Component(NewShellComponent(
		name,
		"/bin/sh",
		[]string{
			"-c",
			strings.Join(append([]string{cmd}, args...), " "),
		}))
}

func (b *builder) BG() Builder {
	b.component = &BGComponent{child: b.component}
	return b
}

func (b *builder) Retry(n int) Builder {
	b.component = &RetryComponent{b.component, n, time.Second}
	return b
}
