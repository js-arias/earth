// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package add implements a command to add pixel values
// to a time pixelation.
package add

import (
	"errors"
	"fmt"
	"os"

	"github.com/js-arias/command"
	"github.com/js-arias/earth"
	"github.com/js-arias/earth/model"
)

var Command = &command.Command{
	Usage: `add [--from <age>] [--to <age>] [--at <age>]
	--in <model-file> --val <value>
	<time-pix-file>`,
	Short: "add pixels to a time pixelation",
	Long: `
Command add reads pixels from a tectonic reconstruction model at a given age,
and them to a time pixelation.

The flag --in is required and is used to provide the name of the input file.

The flag --val is required and sets the value used for the pixels to be
assigned. If the pixel has a value already, the largest value will be stored.

The argument of the command is the file that contains the time pixelation. If
the files does not exist, it will create a new file, if it exists, pixels will
be added to that file.

By default, all time stages of the source model (as defined by --in) will be
used. With the flags --from and --to, it will use only the stages inside of the
indicated ages (in million years). Another possibility is using the flag --at
to set a particular time stage.
	`,
	SetFlags: setFlags,
	Run:      run,
}

var inFlag string
var valFlag int
var fromFlag float64
var toFlag float64
var atFlag float64

func setFlags(c *command.Command) {
	c.Flags().Float64Var(&fromFlag, "from", -1, "")
	c.Flags().Float64Var(&toFlag, "to", -1, "")
	c.Flags().Float64Var(&atFlag, "at", -1, "")
	c.Flags().IntVar(&valFlag, "val", -1, "")
	c.Flags().StringVar(&inFlag, "in", "", "")
}

// MillionYears is used to transform ages in the flags
// (floats in million years)
// to an integer in years.
const millionYears = 1_000_000

func run(c *command.Command, args []string) error {
	if len(args) < 1 {
		return c.UsageError("expecting time pixelation file")
	}
	if valFlag < 0 {
		return c.UsageError("flag --val must be set")
	}
	if inFlag == "" {
		return c.UsageError("flag --in must be set")
	}

	tot, err := readRotModel(inFlag)
	if err != nil {
		return err
	}

	var stages []int64
	if atFlag >= 0 {
		stages = []int64{tot.ClosesStageAge(int64(atFlag * millionYears))}
	} else {
		st := tot.Stages()
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
	}

	output := args[0]
	tp, err := readTimePix(output, tot.Pixelation())
	if err != nil {
		return err
	}
	setTimeValue(tp, tot, stages)

	if err := writeTimePix(output, tp); err != nil {
		return err
	}
	return nil
}

func setTimeValue(tp *model.TimePix, tot *model.Total, ages []int64) {
	for _, a := range ages {
		st := tot.Rotation(a)
		if st == nil {
			continue
		}
		for id := range st {
			tp.Set(a, id, valFlag)
		}
	}
}

func readTimePix(name string, pix *earth.Pixelation) (*model.TimePix, error) {
	f, err := os.Open(name)
	if errors.Is(err, os.ErrNotExist) {
		if pix != nil {
			return model.NewTimePix(pix), nil
		}
		return nil, errors.New("undefined pixelation")
	}
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

func readRotModel(name string) (*model.Total, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// we want an inverse rotation
	// because we are only interested in the stage pixels
	tot, err := model.ReadTotal(f, nil, true)
	if err != nil {
		return nil, fmt.Errorf("when reading file %q: %v", name, err)
	}
	return tot, nil
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
