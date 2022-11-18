// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package mapcmd implements a command to draw
// a time pixelation model as an image map.
package mapcmd

import (
	"encoding/csv"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/js-arias/command"
	"github.com/js-arias/earth/model"
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

	key	color	comment
	0	0, 26, 51	deep ocean
	1	0, 84, 119	oceanic plateaus
	2	68, 167, 196	continental shelf
	3	251, 236, 93	lowlands
	4	255, 165, 0	highlands
	5	229, 229, 224	ice sheets

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
		ages = []int64{tp.CloserStageAge(int64(atFlag * millionYears))}
	} else {
		ages = tp.Stages()
	}

	var keys map[int]color.RGBA
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

func init() {
	rand.Seed(time.Now().UnixNano())
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

func readKey() (map[int]color.RGBA, error) {
	f, err := os.Open(keyFlag)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = '\t'
	r.Comment = '#'

	head, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("while reading file %q: while reading header: %v", keyFlag, err)
	}
	fields := make(map[string]int, len(head))
	for i, h := range head {
		h = strings.ToLower(h)
		fields[h] = i
	}
	for _, h := range []string{"key", "color"} {
		if _, ok := fields[h]; !ok {
			return nil, fmt.Errorf("while reading file %q: expecting field %q", keyFlag, h)
		}
	}

	keys := make(map[int]color.RGBA)
	for {
		row, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		ln, _ := r.FieldPos(0)
		if err != nil {
			return nil, fmt.Errorf("while reading file %q: on row %d: %v", keyFlag, ln, err)
		}

		f := "key"
		k, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf("while reading file %q: on row %d: %v", keyFlag, ln, err)
		}

		f = "color"
		vals := strings.Split(row[fields[f]], ",")
		if len(vals) != 3 {
			return nil, fmt.Errorf("while reading file %q: on row %d: field %q: found %d values", keyFlag, ln, f, len(vals))
		}

		red, err := strconv.Atoi(strings.TrimSpace(vals[0]))
		if err != nil {
			return nil, fmt.Errorf("while reading file %q: on row %d: field %q [red value]: %v", keyFlag, ln, f, err)
		}
		if red > 255 {
			return nil, fmt.Errorf("while reading file %q: on row %d: field %q [red value]: invalid value %d", keyFlag, ln, f, red)
		}

		green, err := strconv.Atoi(strings.TrimSpace(vals[1]))
		if err != nil {
			return nil, fmt.Errorf("while reading file %q: on row %d: field %q [green value]: %v", keyFlag, ln, f, err)
		}
		if green > 255 {
			return nil, fmt.Errorf("while reading file %q: on row %d: field %q [green value]: invalid value %d", keyFlag, ln, f, green)
		}

		blue, err := strconv.Atoi(strings.TrimSpace(vals[2]))
		if err != nil {
			return nil, fmt.Errorf("while reading file %q: on row %d: field %q [blue value]: %v", keyFlag, ln, f, err)
		}
		if blue > 255 {
			return nil, fmt.Errorf("while reading file %q: on row %d: field %q [blue value]: invalid value %d", keyFlag, ln, f, blue)
		}

		c := color.RGBA{uint8(red), uint8(green), uint8(blue), 255}
		keys[k] = c
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("while reading file %q: %v", keyFlag, io.EOF)
	}
	return keys, nil
}

func makeKeyPalette(tp *model.TimePix, ages []int64) map[int]color.RGBA {
	keys := make(map[int]color.RGBA)
	for _, a := range ages {
		for px := 0; px < tp.Pixelation().Len(); px++ {
			v, _ := tp.At(a, px)
			if _, ok := keys[v]; ok {
				continue
			}
			keys[v] = color.RGBA{randUint8(), randUint8(), randUint8(), 255}
		}
	}
	return keys
}

func randUint8() uint8 {
	return uint8(rand.Intn(255))
}

// A stagePix stores a time pixelation
type stagePix struct {
	step float64
	age  int64
	keys map[int]color.RGBA
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

func makeStage(tp *model.TimePix, age int64, keys map[int]color.RGBA) stagePix {
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
