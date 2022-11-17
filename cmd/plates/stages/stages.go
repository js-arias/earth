// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package stages implements a command to print
// the time stages defined in a tectonic reconstruction model.
package stages

import (
	"fmt"
	"os"

	"github.com/js-arias/command"
	"github.com/js-arias/earth/model"
)

var Command = &command.Command{
	Usage: "stages <mode-file>",
	Short: "print time stages of a tectonic reconstruction model",
	Long: `
Command stages reads a tectonic reconstruction model and print the time stages
(in million years) defined in the model.

The first argument of the command is the name of the file that contains the
model.
	`,
	Run: run,
}

// MillionYears is used to transform ages
// an integer in years
// to a float in million years.
const millionYears = 1_000_000

func run(c *command.Command, args []string) error {
	if len(args) < 1 {
		return c.UsageError("expecting tectonic reconstruction model file")
	}

	st, err := readStages(args[0])
	if err != nil {
		return err
	}
	for _, a := range st {
		fmt.Fprintf(c.Stdout(), "%.6f\n", float64(a)/millionYears)
	}
	return nil
}

func readStages(name string) ([]int64, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	tot, err := model.ReadTotal(f, nil, false)
	if err != nil {
		return nil, fmt.Errorf("when reading file %q: %v", name, err)
	}
	return tot.Stages(), nil
}
