// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package values implement a command to print
// the pixel values defined in a time pixelation model.
package values

import (
	"fmt"
	"os"
	"slices"

	"github.com/js-arias/command"
	"github.com/js-arias/earth/model"
)

var Command = &command.Command{
	Usage: "values <time-pix-file>",
	Short: "print pixel values of a time pixelation model",
	Long: `
Command values read a time pixelation model and print the pixel values defined
for the model.

The argument of the command is the name of the file that contains the time
pixelation model.
	`,
	Run: run,
}

func run(c *command.Command, args []string) error {
	if len(args) < 1 {
		return c.UsageError("expecting time pixelation model file")
	}

	pv, err := readValues(args[0])
	if err != nil {
		return err
	}
	for _, v := range pv {
		fmt.Fprintf(c.Stdout(), "%d\n", v)
	}
	return nil
}

func readValues(name string) ([]int, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tp, err := model.ReadTimePix(f, nil)
	if err != nil {
		return nil, fmt.Errorf("when reading file %q: %v", name, err)
	}

	val := make(map[int]bool)
	for _, age := range tp.Stages() {
		s := tp.Stage(age)
		for _, v := range s {
			val[v] = true
		}
	}

	pv := make([]int, 0, len(val))
	for v := range val {
		pv = append(pv, v)
	}
	slices.Sort(pv)

	return pv, nil
}
