// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package euler implements a command to print
// the Euler rotations of a plate.
package euler

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/js-arias/command"
	"github.com/js-arias/earth"
	"github.com/js-arias/earth/rotation"
)

var Command = &command.Command{
	Usage: "euler <rotation-model> [<plate>...]",
	Short: "print Euler rotations of a plate",
	Long: `
Command euler reads a rotation model and prints the rotations for the plates
in that model.

The first argument of the command is the name of the file that contains the
rotation model. One or more plate IDs can be given as additional arguments. If
no plate is given, the command will print the rotations of all plates in the
model.

The output uses the same format as a rotation model:
	
	- The first column is the ID of the moving plate.
	- The second column is the most recent time, in million years.
	- The third column is the latitude of the Euler pole.
	- The fourth column is the longitude of the Euler pole.
	- The fifth column is the angle of the rotation in degrees.
	- The sixth column is the fixed plate.
	`,
	Run: run,
}

func run(c *command.Command, args []string) error {
	if len(args) < 1 {
		return c.UsageError("expecting rotation model file")
	}

	rot, err := readRotationModel(args[0])
	if err != nil {
		return err
	}

	var plates []int
	args = args[1:]
	for _, a := range args {
		p, err := strconv.Atoi(a)
		if err != nil {
			return fmt.Errorf("invalid plate ID %q: %v", a, err)
		}
		plates = append(plates, p)
	}
	if len(args) == 0 {
		plates = rot.Plates()
	}

	for _, p := range plates {
		printEuler(c.Stdout(), rot, p)
	}

	return nil
}

func readRotationModel(name string) (rotation.Rotation, error) {
	f, err := os.Open(name)
	if err != nil {
		return rotation.Rotation{}, err
	}

	rot, err := rotation.Read(f)
	if err != nil {
		return rotation.Rotation{}, fmt.Errorf("when reading file %q: %v", name, err)
	}
	return rot, nil
}

const millionYears = 1_000_000

func printEuler(w io.Writer, rot rotation.Rotation, plate int) {
	e := rot.Euler(plate)
	if len(e) == 0 {
		return
	}

	for _, r := range e {
		t := float64(r.T) / millionYears
		lat := r.E.Latitude()
		lon := r.E.Longitude()
		a := earth.ToDegree(r.Angle)
		fmt.Fprintf(w, "%d\t%.1f\t%.3f\t%.3f\t%.1f\t%d\n", plate, t, lat, lon, a, r.Fix)
	}
}
