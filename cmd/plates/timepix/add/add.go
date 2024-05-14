// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package add implements a command to add pixel values
// to a time pixelation.
package add

import (
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"

	"github.com/js-arias/command"
	"github.com/js-arias/earth"
	"github.com/js-arias/earth/model"
)

var Command = &command.Command{
	Usage: `add [--from <age>] [--to <age>] [--at <age>]
	[-f|--format <format>]
	[--source <value>] [--only <value>] --val <value>
	--in <model-file>
	<time-pix-file>`,
	Short: "add pixels to a time pixelation",
	Long: `
Command add reads pixels from a reconstruction model at a given age, and them
to a time pixelation.

The flag --in is required and is used to provide the name of the input file.
By default, a tectonic reconstruction model will be used, other kind of files
can be used, defined by the flag --format, or -f. Valid formats are:

	mask	 an image used as mask
	model	 default value, a tectonic reconstruction model
	pix	 a pixelated plates file
	timepix  a time pixelation

In the case of a mask image a pixelated plates file, a single time (defined
with the flag --at in million years) must be defined. For the mask image also
it requires that the base time pixelation exists. The image mask should be in
plate carrée projection (also known as equirectangular projection), and only
pixels in white will be set with the indicated value.

In the case of a time pixelation, the value of the source pixelation is the
same as the value to be added. To read a different source pixel value, use the
--source flag.

The flag --val is required and sets the value used for the pixels to be
assigned. If the pixel has a value already, the largest value will be stored.
With the flag --only, only the pixels defined with the given value in the
destination pixelation will be modified.

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
var format string
var valFlag int
var srcFlag int
var onlyFlag int
var fromFlag float64
var toFlag float64
var atFlag float64

func setFlags(c *command.Command) {
	c.Flags().Float64Var(&fromFlag, "from", -1, "")
	c.Flags().Float64Var(&toFlag, "to", -1, "")
	c.Flags().Float64Var(&atFlag, "at", -1, "")
	c.Flags().IntVar(&onlyFlag, "only", -1, "")
	c.Flags().IntVar(&srcFlag, "source", -1, "")
	c.Flags().IntVar(&valFlag, "val", -1, "")
	c.Flags().StringVar(&format, "format", "model", "")
	c.Flags().StringVar(&format, "f", "model", "")
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
	output := args[0]

	var tp *model.TimePix

	if format == "" {
		format = "model"
	}
	switch strings.ToLower(format) {
	case "model":
		tot, err := readRotModel(inFlag)
		if err != nil {
			return err
		}

		var stages []int64
		if atFlag >= 0 {
			stages = []int64{tot.ClosestStageAge(int64(atFlag * millionYears))}
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

		tp, err = readTimePix(output, tot.Pixelation())
		if err != nil {
			return err
		}
		setTimeValue(tp, tot, stages)
	case "mask":
		if atFlag < 0 {
			return fmt.Errorf("flag --at must be set for an image map")
		}
		age := int64(atFlag * millionYears)

		mask, err := readMask(inFlag)
		if err != nil {
			return err
		}

		tp, err = readTimePix(output, nil)
		if err != nil {
			return err
		}

		setMaskValue(tp, mask, age)
	case "pix":
		if atFlag < 0 {
			return fmt.Errorf("flag --at must be set for an image map")
		}
		age := int64(atFlag * millionYears)

		pp, err := readPixPlate(inFlag)
		if err != nil {
			return err
		}

		tp, err = readTimePix(output, pp.Pixelation())
		if err != nil {
			return err
		}
		setPixValue(tp, pp, age)
	case "timepix":
		src, err := readSourceTimePix(inFlag)
		if err != nil {
			return err
		}

		var stages []int64
		if atFlag >= 0 {
			stages = []int64{src.ClosestStageAge(int64(atFlag * millionYears))}
		} else {
			st := src.Stages()
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

		if srcFlag <= 0 {
			srcFlag = valFlag
		}

		tp, err = readTimePix(output, src.Pixelation())
		if err != nil {
			return err
		}
		setTimePixValue(tp, src, stages)
	default:
		return fmt.Errorf("format %q, not known", format)
	}

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
			v, _ := tp.At(a, id)
			if onlyFlag > 0 {
				if onlyFlag == v {
					tp.Set(a, id, valFlag)
				}
				continue
			}
			if valFlag > v {
				tp.Set(a, id, valFlag)
			}
		}
	}
}

func setTimePixValue(tp *model.TimePix, src *model.TimePix, ages []int64) {
	for _, a := range ages {
		st := src.Stage(a)
		if st == nil {
			continue
		}
		for id, pv := range st {
			if pv != srcFlag {
				continue
			}
			v, _ := tp.At(a, id)
			if onlyFlag > 0 {
				if onlyFlag == v {
					tp.Set(a, id, valFlag)
				}
				continue
			}
			if valFlag > v {
				tp.Set(a, id, valFlag)
			}
		}
	}
}

func setMaskValue(tp *model.TimePix, mask image.Image, age int64) {
	stepX := float64(360) / float64(mask.Bounds().Dx())
	stepY := float64(180) / float64(mask.Bounds().Dy())

	pix := tp.Pixelation()
	for px := 0; px < pix.Len(); px++ {
		v, _ := tp.At(age, px)
		if onlyFlag > 0 {
			if onlyFlag != v {
				continue
			}
		} else if v >= valFlag {
			continue
		}

		pt := pix.ID(px).Point()
		x := int((pt.Longitude() + 180) / stepX)
		y := int((90 - pt.Latitude()) / stepY)
		if r, _, _, _ := mask.At(x, y).RGBA(); r < 1000 {
			continue
		}

		tp.Set(age, px, valFlag)
	}
}

func setPixValue(tp *model.TimePix, pp *model.PixPlate, age int64) {
	for _, p := range pp.Plates() {
		for _, id := range pp.Pixels(p) {
			px := pp.Pixel(p, id)
			if px.Begin < age || px.End > age {
				continue
			}
			v, _ := tp.At(age, id)
			if onlyFlag > 0 {
				if onlyFlag == v {
					tp.Set(age, id, valFlag)
				}
				continue
			}
			if valFlag > v {
				tp.Set(age, id, valFlag)
			}
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

func readSourceTimePix(name string) (*model.TimePix, error) {
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

func readMask(name string) (image.Image, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("when decoding image mask %q: %v", name, err)
	}
	return img, nil
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
