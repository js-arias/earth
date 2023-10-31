// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package plates implements a command to print
// the plate IDs defined for a rotation model.
package plates

import (
	"fmt"
	"os"

	"github.com/js-arias/command"
	"github.com/js-arias/earth/rotation"
)

var Command = &command.Command{
	Usage: "plates <rotation-model>",
	Short: "print plates defined for the model",
	Long: `
Command plates reads a rotation model and prints the plate IDs of the plates
defined in the model.

The first argument of the command is the name of the file that contains the
rotation model.
	`,
	Run: run,
}

func run(c *command.Command, args []string) error {
	if len(args) < 1 {
		return c.UsageError("expecting rotation model file")
	}

	p, err := readRotationPlates(args[0])
	if err != nil {
		return err
	}
	for _, id := range p {
		fmt.Fprintf(c.Stdout(), "%d\n", id)
	}
	return nil
}

func readRotationPlates(name string) ([]int, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	rot, err := rotation.Read(f)
	if err != nil {
		return nil, fmt.Errorf("when reading file %q: %v", name, err)
	}
	return rot.Plates(), nil
}
