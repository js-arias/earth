// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// PlatesGUI is a graphic tool
// to manipulate paleogeographic reconstruction models.
package main

import (
	"github.com/js-arias/command"
	"github.com/js-arias/earth/cmd/platesgui/timepix"
)

var app = &command.Command{
	Usage: "platesgui <command> [<argument>...]",
	Short: "a graphic tool for paleogeographic reconstructions",
}

func init() {
	app.Add(timepix.Command)
}

func main() {
	app.Main()
}
