// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package stat provides general statistical functions.
package stat

import (
	"github.com/js-arias/earth"
	"github.com/js-arias/earth/model"
	"github.com/js-arias/earth/stat/pixprob"
	"golang.org/x/exp/slices"
)

// DistProber is an interface for a discrete spherical PDF
// that is only defined by the distance
// (i.e. is isotropic).
type DistProber interface {
	// Prob returns the value of the probability density function
	// for a pixel at a distance dist
	// (in radians).
	Prob(float64) float64
}

type pixDensity struct {
	pix  int
	prob float64
}

// KDE implements a Kernel Density Estimation
// using distribution d as the kernel,
// a set of weighted points p
// (a map of pixel IDs to weight*value of the pixel),
// a time pixelation
// the age of the destination raster,
// a set of pixel priors.
// It return pixel values scaled to their CDF.
func KDE(d DistProber, p map[int]float64, tp *model.TimePix, age int64, prior pixprob.Pixel) map[int]float64 {
	age = tp.ClosestStageAge(age)

	// calculates the raw density of all pixels
	var cum float64
	raw := make([]pixDensity, 0, tp.Pixelation().Len())
	for px := 0; px < tp.Pixelation().Len(); px++ {
		v, _ := tp.At(age, px)
		pp := 1.0
		if prior != nil {
			pp = prior.Prior(v)
			if pp == 0 {
				continue
			}
		}

		pt1 := tp.Pixelation().ID(px).Point()

		var sum float64
		for rp, w := range p {
			pt2 := tp.Pixelation().ID(rp).Point()
			dist := earth.Distance(pt1, pt2)
			sum += d.Prob(dist) * w
		}
		if sum == 0 {
			continue
		}
		p := sum * pp
		raw = append(raw, pixDensity{
			pix:  px,
			prob: p,
		})
		cum += p
	}

	// scale values
	slices.SortFunc(raw, func(a, b pixDensity) bool {
		// descending sort
		return a.prob > b.prob
	})
	cdf := cum
	density := make(map[int]float64, len(raw))
	for _, r := range raw {
		density[r.pix] = cdf / cum
		cdf -= r.prob
	}

	return density
}
