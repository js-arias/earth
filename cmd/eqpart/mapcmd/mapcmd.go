// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package mapcmd implements a command to draw a pixelation
// as an image map.
package mapcmd

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/js-arias/blind"
	"github.com/js-arias/command"
	"github.com/js-arias/earth"
	"github.com/js-arias/earth/vector"
)

var Command = &command.Command{
	Usage: `map [-e|--equator <value>] [-c|--columns <value>]
	[--box <lat,lon,lat,lon>] [--mask <image>]
	[--points] [--pixels] [--random <value>]
	[--bg <image>] -o|--output <out-img-file>`,
	Short: "draw a map of a pixelation",
	Long: `
Package map draws the pixels of pixelation based on an equal area partitioning
of a sphere into an image file using a plate carrée (equirectangular)
projection.

By default the pixelation will have 360 pixels in the equator. Use the flag
--equator, or -e, to change the size of the pixelation.

The flag --output, or -o, is required, and indicates the name of the file of
the output image. In the image each pixel in the equal area pixelation will
have the same color (selected at random). By default the image will be 3600
pixels wide, use the flag --column, or -c, to define a different number of
image columns.

If the flag --bg is defined, the read image file will be used as the background
image, so the pixel colors will be taken from that image.

If the flag --box is defined, only pixels inside the box will be draw. The box
is defined using the format "lat,lon,lat,lon", for example "14,-94,-58,-26"
will enclose South America.

If the flag --mask is defined, the read image file will be used as a mask, so
only pixels that are white in the mask will be draw. This flag can be combined
with --box.

If the flag --points is defined, one or more coordinate points will be read
from the standard input. One coordinate is read per line (each coordinate
separated by one or more spaces), first latitude and the longitude. Lines
starting with '#' will be ignored. If the flag --pixels is defined, the input
values will be interpreted as pixel IDs (one ID per line). The points will be
drawn in solid red (RGB = 255, 0, 0) so, hopefully, they will be easy
identified in the resulting image.

If the flag --random is defined, the indicated number of random pixels will be
added. The pixels will be in solid red (RGB = 255, 0, 0).
	`,
	SetFlags: setFlags,
	Run:      run,
}

var colsFlag int
var equator int
var randFlag int
var boxFlag string
var bgFile string
var maskFile string
var output string
var points bool
var pixFlag bool

func setFlags(c *command.Command) {
	c.Flags().BoolVar(&points, "points", false, "")
	c.Flags().BoolVar(&pixFlag, "pixels", false, "")
	c.Flags().IntVar(&colsFlag, "columns", 3600, "")
	c.Flags().IntVar(&colsFlag, "c", 3600, "")
	c.Flags().IntVar(&equator, "equator", 360, "")
	c.Flags().IntVar(&equator, "e", 360, "")
	c.Flags().IntVar(&randFlag, "random", 0, "")
	c.Flags().StringVar(&bgFile, "bg", "", "")
	c.Flags().StringVar(&boxFlag, "box", "", "")
	c.Flags().StringVar(&maskFile, "mask", "", "")
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

	var boxMask *box
	if boxFlag != "" {
		var err error
		boxMask, err = getBox()
		if err != nil {
			return err
		}
	}

	var maskImage image.Image
	if maskFile != "" {
		var err error
		maskImage, err = readImage(maskFile)
		if err != nil {
			return err
		}
	}

	pix := earth.NewPixelation(equator)
	var img *mapImg
	if bgFile != "" {
		bg, err := readImage(bgFile)
		if err != nil {
			return err
		}
		img = makeBgImage(pix, bg, maskImage, boxMask)
	} else {
		img = makeRndImage(pix, maskImage, boxMask)
	}

	if pixFlag {
		ids, err := inPixels(c.Stdin(), pix.Len())
		if err != nil {
			return err
		}
		for _, id := range ids {
			img.set(id, color.RGBA{255, 0, 0, 255})
		}
	} else if points {
		pts, err := inLatLon(c.Stdin())
		if err != nil {
			return err
		}

		for _, pt := range pts {
			id := pix.Pixel(pt.Lat, pt.Lon).ID()
			img.set(id, color.RGBA{255, 0, 0, 255})
		}
	}
	if randFlag > 0 {
		for i := 0; i < randFlag; i++ {
			id := pix.Random().ID()
			img.set(id, color.RGBA{255, 0, 0, 255})
		}
	}

	if err := writeImage(output, img); err != nil {
		return err
	}
	return nil
}

type mapImg struct {
	step  float64
	color map[int]color.RGBA
	pix   *earth.Pixelation
}

func (m *mapImg) ColorModel() color.Model { return color.RGBAModel }
func (m *mapImg) Bounds() image.Rectangle { return image.Rect(0, 0, colsFlag, colsFlag/2) }
func (m *mapImg) At(x, y int) color.Color {
	lat := 90 - float64(y)*m.step
	lon := float64(x)*m.step - 180

	pos := m.pix.Pixel(lat, lon).ID()
	c, ok := m.color[pos]
	if !ok {
		return color.RGBA{0, 0, 0, 0}
	}
	return c
}

func (m *mapImg) set(px int, c color.RGBA) {
	m.color[px] = c
}

func makeBgImage(pix *earth.Pixelation, bg, mask image.Image, boxMask *box) *mapImg {
	img := &mapImg{
		step:  360 / float64(colsFlag),
		color: make(map[int]color.RGBA, pix.Len()),
		pix:   pix,
	}

	stepX := float64(360) / float64(bg.Bounds().Dx())
	stepY := float64(180) / float64(bg.Bounds().Dy())
	var maskX, maskY float64
	if mask != nil {
		maskX = float64(360) / float64(mask.Bounds().Dx())
		maskY = float64(180) / float64(mask.Bounds().Dy())
	}
	for id := 0; id < pix.Len(); id++ {
		px := pix.ID(id).Point()
		if boxMask != nil {
			if !boxMask.isInside(px.Latitude(), px.Longitude()) {
				continue
			}
		}
		if mask != nil {
			x := int((px.Longitude() + 180) / maskX)
			y := int((90 - px.Latitude()) / maskY)
			r, _, _, a := mask.At(x, y).RGBA()
			if (a>>8) < 200 || (r>>8) < 200 {
				continue
			}
		}
		x := int((px.Longitude() + 180) / stepX)
		y := int((90 - px.Latitude()) / stepY)

		r, g, b, a := bg.At(x, y).RGBA()
		c := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
		img.color[id] = c
	}

	return img
}

func makeRndImage(pix *earth.Pixelation, mask image.Image, boxMask *box) *mapImg {
	img := &mapImg{
		step:  360 / float64(colsFlag),
		color: make(map[int]color.RGBA, pix.Len()),
		pix:   pix,
	}

	var maskX, maskY float64
	if mask != nil {
		maskX = float64(360) / float64(mask.Bounds().Dx())
		maskY = float64(180) / float64(mask.Bounds().Dy())
	}
	for id := 0; id < pix.Len(); id++ {
		px := pix.ID(id).Point()
		if boxMask != nil {
			if !boxMask.isInside(px.Latitude(), px.Longitude()) {
				continue
			}
		}

		if mask != nil {
			x := int((px.Longitude() + 180) / maskX)
			y := int((90 - px.Latitude()) / maskY)
			r, _, _, a := mask.At(x, y).RGBA()
			if (a>>8) < 200 || (r>>8) < 200 {
				continue
			}
		}

		img.color[id] = randColor()
	}
	return img
}

func randColor() color.RGBA {
	return blind.Sequential(blind.Iridescent, rand.Float64())
}

func readImage(name string) (image.Image, error) {
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

func inLatLon(in io.Reader) ([]vector.Point, error) {
	var pts []vector.Point

	r := bufio.NewReader(in)
	for i := 1; ; i++ {
		ln, err := r.ReadString('\n')
		if ln == "" && err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("at line %d: %v", i, err)
		}

		if ln == "" {
			continue
		}
		if ln[0] == '#' {
			continue
		}
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		v := strings.Fields(ln)
		if len(v) < 2 {
			return nil, fmt.Errorf("at line %d: invalid value %q: expecting \"lat lon\"", i, ln)
		}
		pt, err := vector.ParsePoint(v[0], v[1])
		if err != nil {
			return nil, fmt.Errorf("at line %d: %v", i, err)
		}
		pts = append(pts, pt)
	}
	return pts, nil
}

func inPixels(in io.Reader, max int) ([]int, error) {
	var ids []int

	r := bufio.NewReader(in)
	for i := 1; ; i++ {
		ln, err := r.ReadString('\n')
		if ln == "" && err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("at line %d: %v", i, err)
		}

		if ln == "" {
			continue
		}
		if ln[0] == '#' {
			continue
		}
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}

		id, err := readPixID(ln, max)
		if err != nil {
			return nil, fmt.Errorf("at line %d: %v", i, err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func readPixID(s string, max int) (int, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid value %q: %v", s, err)
	}
	if v >= max {
		return 0, fmt.Errorf("invalid value %q: invalid pixel", s)
	}
	return v, nil
}

type box struct {
	p1 earth.Point
	p2 earth.Point
}

func getBox() (*box, error) {
	cs := strings.Split(boxFlag, ",")
	if len(cs) != 4 {
		return nil, fmt.Errorf("invalid --box value %q", boxFlag)
	}

	p1, err := parsePoint(cs[0], cs[1])
	if err != nil {
		return nil, err
	}
	p2, err := parsePoint(cs[2], cs[3])
	if err != nil {
		return nil, err
	}
	if p1.Latitude() < p2.Latitude() {
		p1, p2 = earth.NewPoint(p2.Latitude(), p1.Longitude()), earth.NewPoint(p1.Latitude(), p2.Longitude())
	}
	if p1.Longitude() > p2.Longitude() {
		p1, p2 = earth.NewPoint(p1.Latitude(), p2.Longitude()), earth.NewPoint(p2.Latitude(), p1.Longitude())
	}

	return &box{
		p1: p1,
		p2: p2,
	}, nil
}

func (b *box) isInside(lat, lon float64) bool {
	if lat > b.p1.Latitude() {
		return false
	}
	if lat < b.p2.Latitude() {
		return false
	}

	if lon < b.p1.Longitude() {
		return false
	}
	if lon > b.p2.Longitude() {
		return false
	}

	return true
}

func parsePoint(c1, c2 string) (earth.Point, error) {
	lat, err := strconv.ParseFloat(c1, 64)
	if err != nil {
		return earth.Point{}, fmt.Errorf("invalid latitude: %v: read %q", err, c1)
	}
	if lat < -90 || lat > 90 {
		return earth.Point{}, fmt.Errorf("invalid latitude: %.6f", lat)
	}

	lon, err := strconv.ParseFloat(c2, 64)
	if err != nil {
		return earth.Point{}, fmt.Errorf("invalid longitude: %v: read %q", err, c2)
	}
	if lon < -180 || lon > 180 {
		return earth.Point{}, fmt.Errorf("invalid longitude: %.6f", lon)
	}

	return earth.NewPoint(lat, lon), nil
}
