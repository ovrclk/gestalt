package main

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/ovrclk/gestalt"
)

func WalkerDev() gestalt.Component {
	g := gestalt.NewGroup("walker-dev")
	g.AddChild(gestalt.SH("cleanup", "echo", "cleaning..."))
	g.AddChild(
		gestalt.SH("start", "while true; do echo .; sleep 1; done").BG()).
		AddChild(g.AddChild(gestalt.SH("ping", "sleep", "1")))
	return g
}

func CreateFarm(name string) gestalt.Component {
	g := gestalt.NewGroup("create-farm")
	g.AddChild(gestalt.SH("create", "echo", "walker farms:create", name))
	g.AddChild(gestalt.SH("test", "sleep", "1"))
	return g
}

func RemoveFarm(name string) gestalt.Component {
	g := gestalt.NewGroup("remove-farm")
	g.AddChild(gestalt.SH("remove", "echo", "walker farms:down", name))
	g.AddChild(gestalt.SH("test", "sleep", "1"))
	return g
}

func FarmSuite() gestalt.Component {
	name := "foo"
	s := gestalt.NewSuite("farms")
	s.AddChild(WalkerDev())
	s.AddChild(CreateFarm(name))
	s.AddChild(RemoveFarm(name))
	return s
}

func LeaderSuite() gestalt.Component {
	s := gestalt.NewSuite("leaders")
	s.AddChild(WalkerDev())

	s.AddChild(gestalt.SH("create", "echo", "waker leaders:up l1")).
		AddChild(gestalt.SH("test", "sleep", "1"))

	s.AddChild(gestalt.SH("create", "echo", "waker leaders:down l1")).
		AddChild(gestalt.SH("test", "sleep", "1"))
	return s
}

func ErrSuite() gestalt.Component {
	s := gestalt.NewSuite("test")
	s.AddChild(gestalt.SH("err", "echo", "hello", "1>&2"))
	return s
}

func ParseSuite() gestalt.Component {
	s := gestalt.NewSuite("parse")

	{
		fn := func(b *bufio.Reader, rctx gestalt.RunCtx) (gestalt.ResultValues, error) {
			vals := make(gestalt.ResultValues)

			line, _, err := b.ReadLine()

			if err != nil {
				return vals, err
			}

			fields := strings.Fields(string(line))

			if len(fields) != 3 {
				return vals, fmt.Errorf("invalid format")
			}

			vals["a"] = fields[0]
			vals["b"] = fields[1]
			vals["c"] = fields[2]

			return vals, nil
		}
		s.AddChild(gestalt.SH("abc", "echo", "foo", "bar", "baz").FN(fn))
	}

	{
		fn := func(b *bufio.Reader, rctx gestalt.RunCtx) (gestalt.ResultValues, error) {
			rctx.Logger().Debugf("captured vals: %v", rctx.Values())
			return nil, nil
		}
		s.AddChild(gestalt.SH("bca", "echo", "whatever").FN(fn))
	}

	return s
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	s := gestalt.NewSuite("walker")
	s.AddChild(ParseSuite())
	/*
		s.AddChild(ErrSuite())
		s.AddChild(FarmSuite())
		s.AddChild(LeaderSuite())
	*/
	if err := gestalt.Run(s); err != nil {
		logrus.Errorf("Error running: %v", err)
	}
}
