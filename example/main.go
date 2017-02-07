package main

import (
	"github.com/ovrclk/gestalt"
	g "github.com/ovrclk/gestalt/builder"
	"github.com/ovrclk/gestalt/component"
)

func DevServer() gestalt.Component {
	return g.Group("dev-server").
		Run(g.SH("cleanup", "echo", "cleanup")).
		Run(g.BG().
			Run(g.SH("start", "while true; do echo .; sleep 1; done"))).
		Run(g.SH("wait", "sleep", "10")).
		Run(g.Retry(5).
			Run(g.SH("check", "echo", "check")))
}

// create a new farm
func GroupUp() gestalt.Component {
	check := g.SH("check", "echo", "a", "{{group-name}}-id", "b").
		FN(g.P().Capture("_", "group-host", "_")).
		WithMeta(g.Export("group-host"))
	return g.Group("group-up").
		Run(
			g.SH("create", "echo", "create-group", "{{group-name}}")).
		Run(
			g.Retry(60 * 5).
				Run(check)).
		WithMeta(g.Require("group-name").Export("group-host"))
}

// teardown farm
func GroupDown() gestalt.Component {
	return g.Group("group-down").
		Run(g.SH("down", "echo", "group:down", "{{group-name}}")).
		WithMeta(g.Require("group-name"))
}

// create farm, ensure deletion after child components run
func GroupUpDown() component.Ensure {
	return g.Ensure("groups").
		First(GroupUp()).
		Finally(GroupDown())
}

// create leader on farm host.
func UserUp() gestalt.Component {
	create := g.SH("create",
		"{{group-host}}", "user:create", "{{user-name}}")

	get := g.SH("get", "echo", "{{group-host}}-user").
		FN(g.
			Columns("user-host").
			EnsureCount(1).
			Capture("user-host")).
		WithMeta(g.Export("user-host"))

	check := g.SH("check",
		"echo", "ping", "{{user-host}}")

	return g.Group("user-up").
		Run(create).
		Run(
			g.Retry(60 * 5).
				Run(get)).
		Run(
			g.Retry(60 * 5).
				Run(check)).
		WithMeta(g.
			Require("group-host", "user-name").
			Export("leader-host"))
}

// destroy leader.
func UserDown() gestalt.Component {
	delete := g.SH("delete",
		"echo", "delete", "{{group-host}}", "user:down", "{{user-name}}")
	return g.Group("user-down").
		Run(delete).
		WithMeta(g.Require("group-host", "user-name"))
}

// create then ensure destruction of leader.
func UserUpDown() component.Ensure {
	return g.Ensure("users").
		First(UserUp()).
		Finally(UserDown())
}

func Suite() gestalt.Component {
	return g.Suite("app").
		Run(DevServer()).
		Run(GroupUpDown().
			Run(UserUpDown())).
		WithMeta(g.Require("group-name", "user-name"))
}

func main() {
	gestalt.Run(Suite().
		WithMeta(g.
			Default("group-name", "g1").
			Default("user-name", "u1")))
}
