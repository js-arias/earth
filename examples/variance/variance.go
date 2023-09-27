// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Variance plots the variance of an spherical normal
// relative to the concentration parameter
// used to define the spherical normal.
package main

import (
	"flag"
	"image/color"

	"github.com/js-arias/earth"
	"github.com/js-arias/earth/stat/dist"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

var equator int
var samples int
var maxLambda float64

func init() {
	flag.IntVar(&equator, "equator", 360, "number of pixels at equator (default 360)")
	flag.IntVar(&samples, "samples", 1000, "number of samples for calculation of variance")
	flag.Float64Var(&maxLambda, "max", 10, "maximum value of lambda")
}

func main() {
	flag.Parse()
	pix := earth.NewPixelation(equator)

	p := plot.New()
	p.Title.Text = "variance vs lambda"
	p.X.Label.Text = "lambda (radians^-2)"
	p.Y.Label.Text = "variance (radians^2)"

	ln, err := plotter.NewLine(lambdaVar(pix))
	if err != nil {
		panic(err)
	}
	ln.LineStyle.Width = vg.Points(3)
	ln.LineStyle.Color = color.RGBA{255, 0, 0, 255}
	p.Add(ln)

	p.Y.Min = 0

	// save the plot to a PNG file
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "pix-dist.png"); err != nil {
		panic(err)
	}
}

func lambdaVar(pix *earth.Pixelation) plotter.XYs {
	v := make(plotter.XYs, 1000)
	delta := maxLambda / 1000
	for i := 0; i < 1000; i++ {
		lambda := float64(i)*delta + delta/2
		n := dist.NewNormal(lambda, pix)
		v[i].X = lambda
		v[i].Y = n.Variance(samples)
	}

	return v
}
