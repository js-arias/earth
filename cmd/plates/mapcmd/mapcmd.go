// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package mapcmd implements a command to draw
// a plate motion model as an image map.
package mapcmd

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"

	"github.com/js-arias/blind"
	"github.com/js-arias/command"
	"github.com/js-arias/earth"
	"github.com/js-arias/earth/model"
)

var Command = &command.Command{
	Usage: `map [-c|--columns <value>] [--at <age>]
	-o|--output <out-image-file> <model-file>`,
	Short: "draw a map from a plate motion model",
	Long: `
Command map reads a plate motion model and draw the reconstruction at the
indicated time stage as a png image, using a plate carrée projection.

The argument of the command is the name of the file that contains the plate
motion model.

The flag --output, or -o, is required and sets the name of the output image. If
multiple stages are used, the time stage will append to the name of the image.
In the image all pixels of a given plate will have the same color (selected at
random). By default the image will be 3600 pixels wide, use the flag --columns,
or -c, to define a different number of image columns.

By default all time stages will be produced. Use the flag --at to define a
particular time stage to be draw (in million years).
	`,
	SetFlags: setFlags,
	Run:      run,
}

var colsFlag int
var atFlag float64
var output string

func setFlags(c *command.Command) {
	c.Flags().IntVar(&colsFlag, "columns", 3600, "")
	c.Flags().IntVar(&colsFlag, "c", 3600, "")
	c.Flags().Float64Var(&atFlag, "at", -1, "")
	c.Flags().StringVar(&output, "output", "", "")
	c.Flags().StringVar(&output, "o", "", "")
}

// MillionYears is used to transform ages
// (a float in million years)
// to an integer in years.
const millionYears = 1_000_000

func run(c *command.Command, args []string) error {
	if len(args) == 0 {
		return c.UsageError("expecting plate motion model file")
	}
	if output == "" {
		return c.UsageError("undefined output image flag --output")
	}

	rec, err := readRecons(args[0])
	if err != nil {
		return err
	}
	var ages []int64
	if atFlag >= 0 {
		ages = []int64{int64(atFlag * millionYears)}
	} else {
		ages = rec.Stages()
	}

	pc := makePlatePalette(rec)

	for _, a := range ages {
		name := fmt.Sprintf("%s-%d.png", output, a/millionYears)
		if err := writeImage(name, makeStage(rec, a, pc)); err != nil {
			return err
		}
	}
	return nil
}

func readRecons(name string) (*model.Recons, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	rec, err := model.ReadReconsTSV(f, nil)
	if err != nil {
		return nil, fmt.Errorf("when reading file %q: %v", name, err)
	}
	return rec, nil
}

// A stageModel stores the pixelation of a reconstruction.
type stageModel struct {
	step   float64
	color  map[int]color.RGBA
	pix    *earth.Pixelation
	plates map[int]int
}

func (s stageModel) ColorModel() color.Model { return color.RGBAModel }
func (s stageModel) Bounds() image.Rectangle { return image.Rect(0, 0, colsFlag, colsFlag/2) }
func (s stageModel) At(x, y int) color.Color {
	lat := 90 - float64(y)*s.step
	lon := float64(x)*s.step - 180

	pix := s.pix.Pixel(lat, lon).ID()
	p, ok := s.plates[pix]
	if !ok {
		return color.RGBA{153, 153, 153, 255}
	}
	return s.color[p]
}

func makeStage(rec *model.Recons, age int64, pc map[int]color.RGBA) stageModel {
	plates := make(map[int]int)

	for _, p := range rec.Plates() {
		sp := rec.PixStage(p, age)
		for _, ids := range sp {
			for _, id := range ids {
				plates[id] = p
			}
		}
	}

	return stageModel{
		step:   360 / float64(colsFlag),
		color:  pc,
		pix:    rec.Pixelation(),
		plates: plates,
	}
}

func makePlatePalette(rec *model.Recons) map[int]color.RGBA {
	plates := rec.Plates()
	pc := make(map[int]color.RGBA, len(plates))
	for _, plate := range plates {
		pc[plate] = randColor()
	}
	return pc
}

func randColor() color.RGBA {
	return blind.Sequential(blind.Iridescent, rand.Float64())
}

func writeImage(name string, sm stageModel) (err error) {
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

	if err := png.Encode(f, sm); err != nil {
		return fmt.Errorf("when encoding image file %q: %v", name, err)
	}
	return nil
}
