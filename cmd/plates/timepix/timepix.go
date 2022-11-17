// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package timepix is a metapackage for commands
// that dealt with time pixelations.
package timepix

import (
	"github.com/js-arias/command"
	"github.com/js-arias/earth/cmd/plates/timepix/add"
)

var Command = &command.Command{
	Usage: "timepix <command> [<argument>...]",
	Short: "commands for time pixelation files",
}

func init() {
	Command.Add(add.Command)
}
