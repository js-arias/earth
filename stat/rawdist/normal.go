// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package rawdist provide discretized spherical distribution types,
// that are un-normalized,
// that is,
// they do not sum to one,
// and its the responsibility of the caller to make
// the appropriate integration.
package rawdist

import "math"

// Normal is an unscaled discretization
// of an spherical normal distribution.
//
// It is based on equation (2) from
// Hauberg (2018) 2018 IEEE FUSION: 704
// <doi:10.23919/ICIF.2018.8455242>:
//
//	SN(x|u,v) ∝ exp(-λ * gcd(x,u)^2/2)
//
// where x and u are points on a sphere,
// u is the mean,
// λ is the concentration parameter
// (in 1/radian^2),
// and gcd is the great circle distance.
type Normal struct {
	lambda float64
	step   float64

	pdf    []float64
	logPDF []float64
}

// NewNormal returns an unscaled and discretized
// spherical normal distribution,
// using lambda as the concentration parameter
// (in 1/radian^2)
// using scale as the discretization scale for the distance.
func NewNormal(lambda float64, scale int) Normal {
	step := math.Pi / float64(scale)
	logPDF := make([]float64, scale+1)
	pdf := make([]float64, scale+1)

	for i := range logPDF {
		dist := float64(i) * step
		log := -lambda * dist * dist / 2
		logPDF[i] = log
		pdf[i] = math.Exp(log)
	}

	return Normal{
		lambda: lambda,
		step:   step,
		pdf:    pdf,
		logPDF: logPDF,
	}
}

// Lambda returns the concentration parameter (in 1/radian^2)
// of the spherical normal distribution.
func (n Normal) Lambda() float64 {
	return n.lambda
}

// LogProb returns the natural logarithm
// of the probability density function
// at a distance dist (in radian).
func (n Normal) LogProb(dist float64) float64 {
	r := int(math.Round(dist / n.step))
	if r >= len(n.logPDF) {
		return n.logPDF[len(n.logPDF)-1]
	}
	return n.logPDF[r]
}

// LogProbRingDist returns the natural logarithm
// of the probability density function
// at a given int scaled distance.
func (n Normal) LogProbRingDist(dist int) float64 {
	return n.logPDF[dist]
}

// Prob returns the value
// of the probability density function
// at a distance dist (in radian).
func (n Normal) Prob(dist float64) float64 {
	r := int(math.Round(dist / n.step))
	if r >= len(n.pdf) {
		return 0
	}
	return n.pdf[r]
}

// ProbRingDist returns the value
// of the probability density function
// at a given int scaled distance.
func (n Normal) ProbRingDist(dist int) float64 {
	return n.pdf[dist]
}

// ScaledProb returns the value
// of the probability density function
// for a pixel at a distance dist (in radian)
// scaled by the maximum probability
// (i.e., by 0 distance).
func (n Normal) ScaledProb(dist float64) float64 {
	return n.Prob(dist)
}

// ScaledProbRingDist returns the value
// of the probability density function
// at a given int scaled distance
// scaled by the maximum probability
// (i.e., by 0 distance).
func (n Normal) ScaledProbRingDist(dist int) float64 {
	return n.ProbRingDist(dist)
}
