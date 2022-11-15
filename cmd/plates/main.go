// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Plates is a tool to manipulate paleogeographic reconstruction models.
package main

import (
	"github.com/js-arias/command"
	"github.com/js-arias/earth/cmd/plates/pixels"
	"github.com/js-arias/earth/cmd/plates/rotate"
)

var app = &command.Command{
	Usage: "plates <command> [<argument>...]",
	Short: "a tool to manipulate paleogeographic reconstruction models",
}

func init() {
	app.Add(pixels.Command)
	app.Add(rotate.Command)
}

func main() {
	app.Main()
}
