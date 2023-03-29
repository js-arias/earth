// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// RandNorm produce random points using an spherical normal.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"time"

	"github.com/js-arias/earth"
	"github.com/js-arias/earth/stat/dist"
)

const equatorPixels = 360
const simulations = 1_000_000
const lambda = 10

func main() {
	pix := earth.NewPixelation(equatorPixels)
	n := dist.NewNormal(lambda, pix)
	pt := pix.Pixel(-26.81, -65.22)

	// simulate the points
	pts := make(map[int]float64, pix.Len())
	t := time.Now()
	for p := 0; p < simulations; p++ {
		px := n.Rand(pt)
		pts[px.ID()]++
	}
	fmt.Fprintf(os.Stdout, "simulation time: %v\n", time.Since(t))

	var max float64
	for _, p := range pts {
		if p > max {
			max = p
		}
	}

	img := probImage{
		step: 360.0 / numCols,
		pix:  pix,
		max:  max,
		pts:  pts,
	}
	if err := writeImage("rnd-pts.png", &img); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

type probImage struct {
	step float64
	pix  *earth.Pixelation

	max float64
	pts map[int]float64
}

const numCols = 3600

func (p *probImage) ColorModel() color.Model { return color.RGBAModel }
func (p *probImage) Bounds() image.Rectangle { return image.Rect(0, 0, numCols, numCols/2) }
func (p *probImage) At(x, y int) color.Color {
	lat := 90 - float64(y)*p.step
	lon := float64(x)*p.step - 180

	pix := p.pix.Pixel(lat, lon).ID()
	v, ok := p.pts[pix]
	if !ok {
		return color.RGBA{255, 255, 255, 255}
	}
	return scaleColor(v / p.max)
}

func scaleColor(scale float64) color.RGBA {
	switch {
	case scale < 0.25:
		g := scale * 4 * 255
		return color.RGBA{0, uint8(g), 255, 255}
	case scale < 0.50:
		b := (scale - 0.25) * 4 * 255
		return color.RGBA{0, 255, 255 - uint8(b), 255}
	case scale < 0.75:
		r := (scale - 0.5) * 4 * 255
		return color.RGBA{uint8(r), 255, 0, 255}
	}
	g := (scale - 0.75) * 4 * 255
	return color.RGBA{255, 255 - uint8(g), 0, 255}
}

func writeImage(name string, img *probImage) (err error) {
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
