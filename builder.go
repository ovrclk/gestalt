package gestalt

import (
	"strings"
	"time"
)

var rootBuilder = &builder{}

type Builder interface {
	Group(string) Builder
	Suite(string) Builder
	Component(Component) Builder
	Run(Builder) Builder
	Build() Component
	Retry(n int) Builder
	BG() Builder
	SH(name string, cmd string, args ...string) Builder
}

type builder struct {
	children  []Builder
	component Component
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

func (b *builder) Group(name string) Builder {
	return b.Component(NewGroup(name))
}

func (b *builder) Suite(name string) Builder {
	return b.Component(NewSuite(name))
}

func (b *builder) Component(c Component) Builder {
	return &builder{component: c}
}

func (b *builder) Build() Component {
	c := b.component
	for _, child := range b.children {
		c.AddChild(child.Build())
	}
	return c
}

func (b *builder) Run(child Builder) Builder {
	b.children = append(b.children, child)
	return b
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
	b.component = &RetryComponent{child: b.component, tries: n, delay: time.Second}
	return b
}
