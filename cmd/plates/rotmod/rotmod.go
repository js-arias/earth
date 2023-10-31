// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package rotmod is a metapackage for commands
// that dealt with rotation models.
package rotmod

import (
	"github.com/js-arias/command"
	"github.com/js-arias/earth/cmd/plates/rotmod/plates"
)

var Command = &command.Command{
	Usage: "rotmod <command> [<argument>...]",
	Short: "commands for rotation models",
}

func init() {
	Command.Add(plates.Command)
}
