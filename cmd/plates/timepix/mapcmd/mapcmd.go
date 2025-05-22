// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package mapcmd implements a command to draw
// a time pixelation model as an image map.
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
	"github.com/js-arias/earth/model"
	"github.com/js-arias/earth/pixkey"
)

var Command = &command.Command{
	Usage: `map [-c|--columns <value>] [--at <age>]
	[--key <key-file>] -o|--output <out-image-file>
	<time-pix-file>`,
	Short: "draw a map from a time pixelation model",
	Long: `
Command map reads a time pixelation model from a file and draws the pixel
values at the indicated time stage as a png image, using a plate carrée
projection.

The argument of the command is the name of the file that contains the time
pixelation model.

The flag --output, or -o, is required and sets the name of the output image. If
multiple stages are used, the time stage will append to the name of the image.
In the image all pixels with a given value will have the same color (selected
at random). With the flag --key a key-file can be used to define the colors to
be used in the output. A key file is a tab-delimited file with the following
required columns:

	key	the value used as identifier
	color	an RGB value separated by commas,
		for example "125,132,148".

Any other column will be ignored. Here is an example of a key file:

	key	color	gray	label
	0	54, 75, 154	255	deep ocean
	1	74, 123, 183	235	oceanic plateaus
	2	152, 202, 225	225	continental shelf
	3	254, 218, 139	195	lowlands
	4	246, 126, 75	185	highlands
	5	231, 231, 231	245	ice sheets

By default the image will be 3600 pixels wide, use the flag --columns, or -c,
to define a different number of image columns.

By default all time stages will be produced. Use the flag --at to define a
particular time stage to be draw (in million years).
	`,
	SetFlags: setFlags,
	Run:      run,
}

var colsFlag int
var atFlag float64
var keyFlag string
var output string

func setFlags(c *command.Command) {
	c.Flags().IntVar(&colsFlag, "columns", 3600, "")
	c.Flags().IntVar(&colsFlag, "c", 3600, "")
	c.Flags().Float64Var(&atFlag, "at", -1, "")
	c.Flags().StringVar(&keyFlag, "key", "", "")
	c.Flags().StringVar(&output, "output", "", "")
	c.Flags().StringVar(&output, "o", "", "")
}

// MillionYears is used to transform ages
// (a float in million years)
// to an integer in years.
const millionYears = 1_000_000

func run(c *command.Command, args []string) error {
	if len(args) == 0 {
		return c.UsageError("expecting time pixelation model")
	}
	if output == "" {
		return c.UsageError("flag --output must be set")
	}

	tp, err := readTimePix(args[0])
	if err != nil {
		return err
	}
	var ages []int64
	if atFlag >= 0 {
		ages = []int64{tp.ClosestStageAge(int64(atFlag * millionYears))}
	} else {
		ages = tp.Stages()
	}

	var keys map[int]color.Color
	if keyFlag != "" {
		keys, err = readKey()
		if err != nil {
			return err
		}
	} else {
		keys = makeKeyPalette(tp, ages)
	}

	for _, a := range ages {
		name := fmt.Sprintf("%s-%d.png", output, a/millionYears)
		if err := writeImage(name, makeStage(tp, a, keys)); err != nil {
			return err
		}
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

func readKey() (map[int]color.Color, error) {
	pk, err := pixkey.Read(keyFlag)
	if err != nil {
		return nil, err
	}

	keys := make(map[int]color.Color)
	for _, k := range pk.Keys() {
		c, ok := pk.Color(k)
		if !ok {
			c = randColor()
		}
		keys[k] = c
	}
	return keys, nil
}

func makeKeyPalette(tp *model.TimePix, ages []int64) map[int]color.Color {
	keys := make(map[int]color.Color)
	for _, a := range ages {
		for px := 0; px < tp.Pixelation().Len(); px++ {
			v, _ := tp.At(a, px)
			if _, ok := keys[v]; ok {
				continue
			}
			keys[v] = randColor()
		}
	}
	return keys
}

func randColor() color.RGBA {
	return blind.Sequential(blind.Iridescent, rand.Float64())
}

// A stagePix stores a time pixelation
type stagePix struct {
	step float64
	age  int64
	keys map[int]color.Color
	tp   *model.TimePix
}

func (s stagePix) ColorModel() color.Model { return color.RGBAModel }
func (s stagePix) Bounds() image.Rectangle { return image.Rect(0, 0, colsFlag, colsFlag/2) }
func (s stagePix) At(x, y int) color.Color {
	lat := 90 - float64(y)*s.step
	lon := float64(x)*s.step - 180

	pix := s.tp.Pixelation().Pixel(lat, lon).ID()
	v, _ := s.tp.At(s.age, pix)
	c, ok := s.keys[v]
	if !ok {
		return color.RGBA{0, 0, 0, 0}
	}
	return c
}

func makeStage(tp *model.TimePix, age int64, keys map[int]color.Color) stagePix {
	return stagePix{
		step: 360 / float64(colsFlag),
		age:  age,
		keys: keys,
		tp:   tp,
	}
}

func writeImage(name string, sp stagePix) (err error) {
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

	if err := png.Encode(f, sp); err != nil {
		return fmt.Errorf("when encoding image file %q: %v", name, err)
	}
	return nil
}
