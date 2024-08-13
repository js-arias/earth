// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// sKDE outputs an spherical KDE.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/js-arias/blind"
	"github.com/js-arias/earth"
	"github.com/js-arias/earth/model"
	"github.com/js-arias/earth/stat"
	"github.com/js-arias/earth/stat/dist"
	"github.com/js-arias/earth/stat/weight"
)

const equatorPixels = 360
const numPoints = 1000
const lambda = 120
const kdeLambda = 1000

func main() {
	pix := earth.NewPixelation(equatorPixels)
	tp := model.NewTimePix(pix)
	for px := 0; px < pix.Len(); px++ {
		tp.Set(0, px, 1)
	}
	pw := weight.New()
	pw.Set(1, 1)

	n := dist.NewNormal(lambda, pix)
	pt := pix.Pixel(-26.81, -65.22)

	// simulate the points
	pts := make(map[int]float64, numPoints)
	for i := 0; i < numPoints; i++ {
		px := n.Rand(pt)
		pts[px.ID()]++
	}

	nKDE := dist.NewNormal(kdeLambda, pix)
	kde := stat.KDE(nKDE, pts, tp, 0, pw)

	img := probImage{
		step: 360.0 / numCols,
		pix:  pix,
		min:  0.05,
		pts:  kde,
	}
	if err := writeImage("kde.png", &img); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

}

type probImage struct {
	step float64
	pix  *earth.Pixelation

	min float64
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
		return color.RGBA{102, 102, 102, 255}
	}
	if v < p.min {
		return color.RGBA{102, 102, 102, 255}
	}
	return blind.Gradient(v)
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
