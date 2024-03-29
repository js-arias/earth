// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package pixels is a metapackage for commands
// that dealt with pixelated plates files.
package pixels

import (
	"github.com/js-arias/command"
	"github.com/js-arias/earth/cmd/plates/pixels/add"
	"github.com/js-arias/earth/cmd/plates/pixels/cat"
	"github.com/js-arias/earth/cmd/plates/pixels/importcmd"
	"github.com/js-arias/earth/cmd/plates/pixels/list"
	"github.com/js-arias/earth/cmd/plates/pixels/mapcmd"
)

var Command = &command.Command{
	Usage: "pixels <command> [<argument>...]",
	Short: "commands for pixelated plates files",
}

func init() {
	Command.Add(add.Command)
	Command.Add(cat.Command)
	Command.Add(importcmd.Command)
	Command.Add(list.Command)
	Command.Add(mapcmd.Command)
}
