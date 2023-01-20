// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package stat provides general statistical functions.
package stat

import (
	"github.com/js-arias/earth"
	"github.com/js-arias/earth/model"
	"github.com/js-arias/earth/stat/pixprob"
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

// QuantileChord2er is an interface for a discrete spherical distribution
// that is isotropic.
type QuantileChord2er interface {
	DistProber

	// QuantileChord2 returns the square of the Euclidean chord distance
	// for the maximum distance
	// that is inside the indicated cumulative density.
	QuantileChord2(float64) float64
}

// KDE implements a Kernel Density Estimation
// using distribution d as the kernel,
// a set of weighted points p
// (a map of pixel IDs to weight*value of the pixel),
// a time pixelation
// the age of the destination raster,
// a set of pixel priors,
// and a bound for the CDF of the distribution
// (i.e. pixels outside the indicated CDF in d will be ignored).
func KDE(d QuantileChord2er, p map[int]float64, tp *model.TimePix, age int64, prior pixprob.Pixel, bound float64) map[int]float64 {
	age = tp.CloserStageAge(age)

	maxChord2 := d.QuantileChord2(bound)
	density := make(map[int]float64)
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
			if earth.Chord2(pt1, pt2) > maxChord2 {
				continue
			}
			dist := earth.Distance(pt1, pt2)
			sum += d.Prob(dist) * w
		}
		if sum == 0 {
			continue
		}
		density[px] = sum * pp
	}
	return density
}
