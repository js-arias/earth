// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package rotate implements a command to add new pixel rotations
// (i.e. pixel locations in the past)
// to a plate motion model.
package rotate

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/js-arias/command"
	"github.com/js-arias/earth"
	"github.com/js-arias/earth/model"
	"github.com/js-arias/earth/rotation"
	"golang.org/x/exp/slices"
)

var Command = &command.Command{
	Usage: `rotate [--from <age>] [--to <age>] [--step <age>]
	--pix <pix-file> --rot <rotation-file>
	<model-file> [<age>...]`,
	Short: "rotate pixels of a plate motion model",
	Long: `
Command rotate reads a rotation file and updates the pixel locations from a
pixelated plate file, and write them into a plate motion model.

The flag --pix is required and sets the file with pixelated plates. The
resolution (i.e. the number of pixels in the equator) of the pixelation must
be equal to the plate motion model (if the model already exists). It will add
any plate or pixel absent in the model.

The flag --rot is required and indicates the file containing a rotation model.
Rotation model files are the standard files for rotations used in tectonic
modelling software such as GPlates.

The first argument of the command is the name of the file that contains the
model. If the file does not exists, it will create a new empty model and store
it in that file.

One or more time stages (in million years) can be given as additional
arguments for the command. Each age is used to set the locations of pixels in
the model to its paleo-location, given the rotation model. If no stages are
defined, the flags --from, --to, and --step, can be used to define the oldest
stage (--from), the most recent stage (--to, default is 0), and the size of
each time interval (--step, default is 5).
	`,
	SetFlags: setFlags,
	Run:      run,
}

var fromFlag float64
var toFlag float64
var stepFlag float64
var pixFile string
var rotFile string

func setFlags(c *command.Command) {
	c.Flags().Float64Var(&fromFlag, "from", 0, "")
	c.Flags().Float64Var(&toFlag, "to", 0, "")
	c.Flags().Float64Var(&stepFlag, "step", 5, "")
	c.Flags().StringVar(&pixFile, "pix", "", "")
	c.Flags().StringVar(&rotFile, "rot", "", "")
}

// MillionYears is used to transform ages
// (a float in million years)
// to an integer in years.
const millionYears = 1_000_000

func run(c *command.Command, args []string) error {
	if len(args) < 1 {
		return c.UsageError("expecting plate motion model file")
	}
	if pixFile == "" {
		return c.UsageError("undefined value for --pix flag")
	}
	if rotFile == "" {
		return c.UsageError("undefined value for --rot flag")
	}

	modFile := args[0]

	args = args[1:]
	var ages []int64
	if len(args) > 0 {
		ages = make([]int64, 0, len(args))
		for _, a := range args {
			v, err := strconv.ParseFloat(a, 64)
			if err != nil {
				msg := fmt.Sprintf("when reading <age> argument %q: %v", a, err)
				return c.UsageError(msg)
			}
			ages = append(ages, int64(v*millionYears))
		}
		slices.Sort(ages)
	} else if fromFlag > toFlag {
		for a := toFlag; a <= fromFlag; a += stepFlag {
			ages = append(ages, int64(a*millionYears))
		}
	} else {
		return c.UsageError("undefined age stages")
	}

	pp, err := readPixPlate(pixFile)
	if err != nil {
		return err
	}
	rot, err := readRotation(rotFile)
	if err != nil {
		return err
	}
	rec, err := readRecons(modFile, pp.Pixelation())
	if err != nil {
		return err
	}

	for _, p := range pp.Plates() {
		for _, a := range ages {
			makeRotation(rec, pp, rot, p, a)
		}
	}

	if err := writeRecons(modFile, rec); err != nil {
		return err
	}

	return nil
}

func readPixPlate(name string) (*model.PixPlate, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pp, err := model.ReadPixPlate(f, nil)
	if err != nil {
		return nil, fmt.Errorf("when reading file %q: %v", name, err)
	}
	return pp, nil
}

func readRotation(name string) (rotation.Rotation, error) {
	f, err := os.Open(name)
	if err != nil {
		return rotation.Rotation{}, err
	}
	defer f.Close()

	rot, err := rotation.Read(f)
	if err != nil {
		return rotation.Rotation{}, fmt.Errorf("when reading file %q: %v", name, err)
	}
	return rot, nil
}

func readRecons(name string, pix *earth.Pixelation) (*model.Recons, error) {
	f, err := os.Open(name)
	if errors.Is(err, os.ErrNotExist) {
		return model.NewRecons(pix), nil
	}
	if err != nil {
		return nil, err
	}

	rec, err := model.ReadReconsTSV(f, pix)
	if err != nil {
		return nil, fmt.Errorf("when reading file %q: %v", name, err)
	}
	return rec, nil
}

func makeRotation(rec *model.Recons, pp *model.PixPlate, rot rotation.Rotation, plate int, age int64) {
	r, ok := rot.Rotation(plate, age)
	if !ok {
		return
	}

	l := pp.Pixels(plate)
	locs := make(map[int][]int, len(l))
	pix := make(map[int]bool, len(l))
	used := make(map[int]bool, len(l))
	first := pp.Pixelation().Len()
	last := 0
	for _, id := range l {
		px := pp.Pixel(plate, id)
		if px.Begin < age || px.End > age {
			continue
		}
		pix[id] = true
		pt := pp.Pixelation().ID(id).Point().Vector()
		v := r.Rotate(pt)
		np := pp.Pixelation().FromVector(v)
		locs[id] = []int{np.ID()}
		used[np.ID()] = true
		if np.ID() < first {
			first = np.ID()
		}
		if np.ID() > last {
			last = np.ID()
		}
	}

	// Get "present" pixels from "past" pixels
	// so we are sure that every pixel in the past
	// has an assignment in the present.
	// This reduce the number of "holes" produced
	// when a rotation is performed
	// because of the discrete nature of the pixelation.
	inv := rotation.Inverse(r)
	for id := first; id <= last; id++ {
		if used[id] {
			continue
		}
		np := pp.Pixelation().ID(id).Point().Vector()
		v := inv.Rotate(np)
		px := pp.Pixelation().FromVector(v)
		if !pix[px.ID()] {
			continue
		}
		locs[px.ID()] = append(locs[px.ID()], id)
	}

	rec.Add(plate, locs, age)
}

func writeRecons(name string, rec *model.Recons) (err error) {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer func() {
		e := f.Close()
		if e != nil && err == nil {
			err = e
		}
	}()

	if err := rec.TSV(f); err != nil {
		return err
	}
	return nil
}
