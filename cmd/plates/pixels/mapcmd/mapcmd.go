// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package mapcmd implements a command to draw a pixelation
// as an image map.
package mapcmd

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"

	"github.com/js-arias/command"
	"github.com/js-arias/earth"
	"github.com/js-arias/earth/model"
	"pgregory.net/rand"
)

var Command = &command.Command{
	Usage: `map [-c|--columns <value>]
	-o|--output <out-img-file> [<pix-file>...]`,
	Short: "draw a map from a file with pixelated plates",
	Long: `
Command map reads one or more pixelated plates files and draw the plates into
a png image using a plate carrée (equirectangular) projection.

The flag --output, or -o, is required, and indicates the name of the file of
the output image. In the image all pixels associated with a plate will have
the same color (selected at random). By default the image will be 3600 pixels
wide, use the flag --column, or -c, to define a different number of image
columns.

One or more input files can be given as arguments. If no file is given the
input will be read from the standard input.
	`,
	SetFlags: setFlags,
	Run:      run,
}

var colsFlag int
var output string

func setFlags(c *command.Command) {
	c.Flags().IntVar(&colsFlag, "columns", 3600, "")
	c.Flags().IntVar(&colsFlag, "c", 3600, "")
	c.Flags().StringVar(&output, "output", "", "")
	c.Flags().StringVar(&output, "o", "", "")
}

func run(c *command.Command, args []string) error {
	if output == "" {
		return c.UsageError("expecting output image file name, flag --output")
	}

	if colsFlag%2 != 0 {
		colsFlag++
	}

	if len(args) == 0 {
		args = append(args, "-")
	}
	var img *mapImg
	for _, a := range args {
		var pix *earth.Pixelation
		if img != nil {
			pix = img.pix
		}

		pp, err := readPixPlate(c.Stdin(), a, pix)
		if err != nil {
			return err
		}
		if img == nil {
			img = &mapImg{
				step:  360 / float64(colsFlag),
				color: make(map[int]color.RGBA),
				pix:   pp.Pixelation(),
				pp:    make(map[int]pixel),
			}
		}
		img.addPixels(pp)
	}

	if img == nil {
		return nil
	}

	if err := writeImage(output, img); err != nil {
		return err
	}
	return nil
}

func readPixPlate(r io.Reader, name string, pix *earth.Pixelation) (*model.PixPlate, error) {
	if name != "-" {
		f, err := os.Open(name)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r = f
	} else {
		name = "stdin"
	}

	pp, err := model.ReadPixPlate(r, pix)
	if err != nil {
		return nil, fmt.Errorf("when reading file %q: %v", name, err)
	}
	return pp, nil
}

type mapImg struct {
	step  float64
	color map[int]color.RGBA
	pix   *earth.Pixelation
	pp    map[int]pixel
}

type pixel struct {
	age   int64
	plate int
}

func (m *mapImg) ColorModel() color.Model { return color.RGBAModel }
func (m *mapImg) Bounds() image.Rectangle { return image.Rect(0, 0, colsFlag, colsFlag/2) }

func (m *mapImg) At(x, y int) color.Color {
	lat := 90 - float64(y)*m.step
	lon := float64(x)*m.step - 180

	pos := m.pix.Pixel(lat, lon).ID()
	pp, ok := m.pp[pos]
	if !ok {
		return color.RGBA{0, 0, 0, 0}
	}
	if c, ok := m.color[pp.plate]; ok {
		return c
	}

	c := color.RGBA{randUint8(), randUint8(), randUint8(), 255}
	m.color[pp.plate] = c
	return c
}

func (m *mapImg) addPixels(pp *model.PixPlate) {
	for _, plate := range pp.Plates() {
		for _, id := range pp.Pixels(plate) {
			px := pp.Pixel(plate, id)
			op, ok := m.pp[id]
			if !ok {
				m.pp[id] = pixel{
					age:   px.Begin,
					plate: plate,
				}
				continue
			}
			if px.Begin < op.age {
				continue
			}
			m.pp[id] = pixel{
				age:   px.Begin,
				plate: plate,
			}
		}
	}
}

func randUint8() uint8 {
	return uint8(rand.Intn(255))
}

func writeImage(name string, img *mapImg) (err error) {
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

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("when encoding image file %q: %v", name, err)
	}
	return nil
}
