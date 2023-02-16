// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// PixDist plots the number of pixels at a given distance
// from a pixel,
// across multiple latitude rings.
//
// This example shows that the pixelation is nearly isotropic.
package main

import (
	"flag"
	"image/color"
	"math"

	"github.com/js-arias/earth"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

var equator int
var perRing int

func init() {
	flag.IntVar(&equator, "equator", 360, "number of pixels at equator (default 360)")
	flag.IntVar(&perRing, "sample", 20, "number of sampled pixels per ring (default 20)")
}

func main() {
	flag.Parse()
	pix := earth.NewPixelation(equator)

	p := plot.New()
	p.Title.Text = "Pixel count"
	p.X.Label.Text = "distance"
	p.Y.Label.Text = "pixels"

	for r := 0; r < pix.Rings(); r++ {
		for i := 0; i < perRing; i++ {
			px := pix.RandInRing(r)
			s, err := plotter.NewScatter(distances(pix, px.ID()))
			if err != nil {
				panic(err)
			}
			s.GlyphStyle.Color = color.RGBA{0, 0, 255, 255}

			p.Add(s)
		}
	}

	// draw the distance from the north pole
	pole, err := plotter.NewLine(distances(pix, 0))
	if err != nil {
		panic(err)
	}
	pole.LineStyle.Width = vg.Points(3)
	pole.LineStyle.Color = color.RGBA{255, 0, 0, 255}
	p.Add(pole)

	// save the plot to a PNG file
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "pix-dist.png"); err != nil {
		panic(err)
	}
}

func distances(pix *earth.Pixelation, px int) plotter.XYs {
	v := make(plotter.XYs, pix.Rings())
	rStep := earth.ToRad(pix.Step())
	for r := 0; r < pix.Rings(); r++ {
		v[r].X = float64(r) * rStep
	}

	pt1 := pix.ID(px).Point()
	for id := 0; id < pix.Len(); id++ {
		pt2 := pix.ID(id).Point()
		dist := earth.Distance(pt1, pt2)

		r := int(math.Round(dist / rStep))
		v[r].Y++
	}

	return v
}
