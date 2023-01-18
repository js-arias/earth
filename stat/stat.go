// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package stat provides general statistical functions.
package stat

import (
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
	Prob(float64)
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
func KDE(d QuantileChord2er, p map[int]float64, tp *model.TimePix, age int64, prior pixprob.Pixel) map[int]float64 {
	return nil
}
