// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package lencmd implements a command to get the number of pixels
// in an isolatitude pixelation.
package lencmd

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strconv"
	"strings"

	"github.com/js-arias/command"
	"github.com/js-arias/earth"
)

var Command = &command.Command{
	Usage: "len [-e|--equator <value>] [--box <lat,lon,lat,lon>] [--rings]",
	Short: "get the number of pixels in a pixelation",
	Long: `
Command len retrieves the number of pixels produced by an isolatitude
pixelation with the given number of pixels at the equator, as well as the
number of isolatitude rings.

By default the pixelation will be of 360 pixels at the equator. Use the flag
--equator, or -e, to define a different pixelation.

If the flag --box is defined, only pixels inside the box will be count. The box
is defined using the format "lat,lon,lat,lon", for example "14,-94,-58,-26"
will enclose South America.

If the flag --mask is defined, the read image file will be used as a mask, so
only pixels that are white in the mask will be count. This flag can be combined
with --box.

If the flag --rings is defined, the number of pixels at each ring will be
printed.
	`,
	SetFlags: setFlags,
	Run:      run,
}

var equator int
var rings bool
var boxFlag string
var maskFile string

func setFlags(c *command.Command) {
	c.Flags().BoolVar(&rings, "rings", false, "")
	c.Flags().IntVar(&equator, "equator", 360, "")
	c.Flags().IntVar(&equator, "e", 360, "")
	c.Flags().StringVar(&boxFlag, "box", "", "")
	c.Flags().StringVar(&maskFile, "mask", "", "")
}

func run(c *command.Command, args []string) error {
	pix := earth.NewPixelation(equator)

	var boxMask *box
	if boxFlag != "" {
		var err error
		boxMask, err = getBox()
		if err != nil {
			return err
		}
	}

	var mask image.Image
	if maskFile != "" {
		var err error
		mask, err = readImage(maskFile)
		if err != nil {
			return err
		}
	}

	if boxMask != nil || mask != nil {
		var maskX, maskY float64
		if mask != nil {
			maskX = float64(360) / float64(mask.Bounds().Dx())
			maskY = float64(180) / float64(mask.Bounds().Dy())
		}

		sum := 0
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
			sum++
		}
		fmt.Fprintf(c.Stdout(), "pixels: %d\ninside mask: %d\n", pix.Len(), sum)
		return nil
	}

	fmt.Fprintf(c.Stdout(), "pixels: %d\nrings: %d\n", pix.Len(), pix.Rings())
	if rings {
		for i := 0; i < pix.Rings(); i++ {
			fmt.Fprintf(c.Stdout(), "ring %d [lat %.6f]: %d\n", i, pix.RingLat(i), pix.PixPerRing(i))
		}
	}

	return nil
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
