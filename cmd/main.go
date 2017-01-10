package main

import "github.com/ovrclk/gestalt"

func main() {
	s := gestalt.NewSuite("foo")

	s.AddChild(gestalt.NewCommandComponent("ls", []string{"/"}))
	s.Run()
}
