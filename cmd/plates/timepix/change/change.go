// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package change implements a command to change pixel values
// of a time pixelation model.
package change

import (
	"fmt"
	"os"
	"slices"

	"github.com/js-arias/command"
	"github.com/js-arias/earth"
	"github.com/js-arias/earth/model"
)

var Command = &command.Command{
	Usage: `change [--from <age>] [--to <age>] [--at <age>]
	--old <value> --new <value> <time-pix-file>`,
	Short: "change pixel values of a time pixelation",
	Long: `
Command change reads a time pixelation model and changes its pixel values.

The flag --old is required and is used to set the pixel value to be changed.
The flag --new is required and is used to set the new value of the pixels.

By default, all time stages of the time pixelation will be changed. With the
flags --from and --to, it will change only the stages inside the indicated
ages (in million years). Another possibility is using the flag --at to change
a particular time stage.

The argument of the command is the file that contains the time pixelation.
This argument is required.
	`,
	SetFlags: setFlags,
	Run:      run,
}

var oldValue int
var newValue int
var fromFlag float64
var toFlag float64
var atFlag float64

func setFlags(c *command.Command) {
	c.Flags().Float64Var(&fromFlag, "from", -1, "")
	c.Flags().Float64Var(&toFlag, "to", -1, "")
	c.Flags().Float64Var(&atFlag, "at", -1, "")
	c.Flags().IntVar(&oldValue, "old", -1, "")
	c.Flags().IntVar(&newValue, "new", -1, "")
}

// MillionYears is used to transform ages in the flags
// (floats in million years)
// to an integer in years.
const millionYears = 1_000_000

func run(c *command.Command, args []string) error {
	if len(args) < 1 {
		return c.UsageError("expecting time pixelation file")
	}

	if oldValue < 0 || newValue < 0 {
		return c.UsageError("flags --old and --new must be defined")
	}

	output := args[0]

	tp, err := readTimePix(output, nil)
	if err != nil {
		return err
	}

	var stages []int64
	if atFlag >= 0 {
		stages = []int64{tp.ClosestStageAge(int64(atFlag * millionYears))}
	} else {
		st := tp.Stages()
		from := st[len(st)-1]
		if fromFlag >= 0 {
			from = int64(fromFlag * millionYears)
		}
		to := st[0]
		if toFlag >= 0 {
			to = int64(toFlag * millionYears)
		}
		stages = make([]int64, 0, len(st))
		for _, a := range st {
			if a > from {
				continue
			}
			if a < to {
				continue
			}
			stages = append(stages, a)
		}
		if len(stages) == 0 {
			return nil
		}
		slices.Sort(stages)
	}

	setTimeValue(tp, stages)

	if err := writeTimePix(output, tp); err != nil {
		return err
	}
	return nil
}

func readTimePix(name string, pix *earth.Pixelation) (*model.TimePix, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tp, err := model.ReadTimePix(f, pix)
	if err != nil {
		return nil, fmt.Errorf("when reading file %q: %v", name, err)
	}
	return tp, nil
}

func setTimeValue(tp *model.TimePix, ages []int64) {
	for _, a := range ages {
		r := tp.Stage(a)
		if r == nil {
			continue
		}
		for pix := 0; pix < tp.Pixelation().Len(); pix++ {
			v, ok := r[pix]
			if !ok {
				continue
			}
			if v != oldValue {
				continue
			}
			tp.Set(a, pix, newValue)
		}
	}
}

func writeTimePix(name string, tp *model.TimePix) (err error) {
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

	if err := tp.TSV(f); err != nil {
		return err
	}
	return nil
}
