// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package rotate implements a command to rotate
// a time pixelation model.
package rotate

import (
	"fmt"
	"os"

	"github.com/js-arias/command"
	"github.com/js-arias/earth"
	"github.com/js-arias/earth/model"
)

var Command = &command.Command{
	Usage: `rotate --model <motion-model> [--unrot]
	-o|--output <time-pix-file> <time-pix-file>`,
	Short: "rotate a time pixelation",
	Long: `
Command rotate reads pixels from a time pixelation file, and a plate motion
model, and rotate the pixels of the time pixelation, to its location at the
defined time stage.

By default, it assumes that the time pixelation has the pixel in the present
locations, and should be moved to past time stages. If the flag --unrot is
defined, the pixels will be assumed to be in a past time stage, and will be
moved to a present time stage.

As it is possible that multiple assignations will be given to a pixel, the
maximum stored value will be preserved.

The time pixelation resulted from the rotation will be stored in the file
indicated by the --output, or -o, flag.

The argument of the command is the file that contains the time pixelation to
be rotated. This argument is required.
	`,
	SetFlags: setFlags,
	Run:      run,
}

var modFile string
var output string
var unRot bool

func setFlags(c *command.Command) {
	c.Flags().StringVar(&modFile, "model", "", "")
	c.Flags().BoolVar(&unRot, "unrot", false, "")
	c.Flags().StringVar(&output, "output", "", "")
	c.Flags().StringVar(&output, "o", "", "")
}

func run(c *command.Command, args []string) error {
	if len(args) < 1 {
		return c.UsageError("expecting time pixelation file")
	}

	if modFile == "" {
		return c.UsageError("flag --model must be defined")
	}
	if output == "" {
		return c.UsageError("flag --output must be defined")
	}

	tp, err := readTimePix(args[0])
	if err != nil {
		return err
	}
	pix := tp.Pixelation()

	tot, err := readRotation(modFile, pix)
	if err != nil {
		return err
	}
	max := lastRotationStage(tot)

	np := model.NewTimePix(pix)
	for _, age := range tp.Stages() {
		if age > max {
			break
		}
		rot := tot.Rotation(age)
		for px := 0; px < pix.Len(); px++ {
			v, _ := tp.At(age, px)
			if v == 0 {
				continue
			}
			dst := rot[px]
			for _, rp := range dst {
				if ov, _ := np.At(age, rp); ov > v {
					continue
				}
				np.Set(age, rp, v)
			}
		}
	}

	if err := writeTimePix(output, np); err != nil {
		return err
	}

	return nil
}

func readTimePix(name string) (*model.TimePix, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tp, err := model.ReadTimePix(f, nil)
	if err != nil {
		return nil, fmt.Errorf("when reading file %q: %v", name, err)
	}
	return tp, nil
}

func readRotation(name string, pix *earth.Pixelation) (*model.Total, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rot, err := model.ReadTotal(f, pix, unRot)
	if err != nil {
		return nil, fmt.Errorf("on file %q: %v", name, err)
	}

	return rot, nil
}

func lastRotationStage(m *model.Total) int64 {
	stages := m.Stages()
	return stages[len(stages)-1]
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
