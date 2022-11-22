// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// IsoLat is a tool to work with isolatitude pixelations.
package main

import (
	"github.com/js-arias/command"
	"github.com/js-arias/earth/cmd/isolat/mapcmd"
	"github.com/js-arias/earth/cmd/isolat/pixel"
)

var app = &command.Command{
	Usage: "isolat <command> [<argument>...]",
	Short: "a toll to work with isolatitude pixelations",
}

func init() {
	app.Add(mapcmd.Command)
	app.Add(pixel.Command)
}

func main() {
	app.Main()
}
