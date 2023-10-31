// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// EqPart is a tool to work with a pixelation based on an equal area partitioning.
package main

import (
	"github.com/js-arias/command"
	"github.com/js-arias/earth/cmd/eqpart/ids"
	"github.com/js-arias/earth/cmd/eqpart/lencmd"
	"github.com/js-arias/earth/cmd/eqpart/mapcmd"
	"github.com/js-arias/earth/cmd/eqpart/pixel"
	"github.com/js-arias/earth/cmd/eqpart/variance"
)

var app = &command.Command{
	Usage: "eqpart <command> [<argument>...]",
	Short: "a tool to work with pixelation based on an equal area partitioning",
}

func init() {
	app.Add(ids.Command)
	app.Add(lencmd.Command)
	app.Add(mapcmd.Command)
	app.Add(pixel.Command)
	app.Add(variance.Command)
}

func main() {
	app.Main()
}
