// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package dist provides spherical distribution types.
package dist

import (
	"math"

	"github.com/js-arias/earth"
)

// Normal is an isotropic univariate spherical normal distribution
// discretized over a pixelation.
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
// (in 1/radians),
// and gcd is the great circle distance.
type Normal struct {
	pix    *earth.Pixelation
	step   float64 // step of a ring in radians
	lambda float64 // concentration parameter

	pdf    []float64
	cdf    []float64
	ring   []float64
	logPDF []float64
}

// NewNormal returns a discretized spherical normal,
// using lambda as the concentration parameter
// (in 1/radian units)
// and using pix as the underlying pixelation.
func NewNormal(lambda float64, pix *earth.Pixelation) Normal {
	rings := pix.Rings()
	logPDF := make([]float64, rings)
	cdf := make([]float64, rings)
	ring := make([]float64, rings)

	rStep := earth.ToRad(pix.Step())

	// get initial values
	var sum float64
	for i := range logPDF {
		dist := float64(i) * rStep
		logP := -lambda * dist * dist / 2
		logPDF[i] = logP

		logR := logP + math.Log(float64(pix.PixPerRing(i)))
		pRing := math.Exp(logR)
		ring[i] = pRing
		sum += pRing
		cdf[i] = sum
	}

	// scale values
	pdf := make([]float64, rings)
	logSum := math.Log(sum)
	for i := range logPDF {
		ring[i] = ring[i] / sum
		cdf[i] = cdf[i] / sum
		logPDF[i] = logPDF[i] - logSum
		pdf[i] = math.Exp(logPDF[i])
	}

	return Normal{
		pix:    pix,
		step:   rStep,
		lambda: lambda,
		pdf:    pdf,
		cdf:    cdf,
		ring:   ring,
		logPDF: logPDF,
	}
}

// CDF returns the probability cumulative density function
// for a pixel at a distance dist
// (in radians).
func (n Normal) CDF(dist float64) float64 {
	r := int(math.Round(dist / n.step))
	if r >= len(n.cdf) {
		return 1
	}
	return n.cdf[r]
}

// Lambda returns the concentration parameter
// (in 1/radians)
// of a normal distribution.
func (n Normal) Lambda() float64 {
	return n.lambda
}

// LogProb returns the natural logarithm
// of the probability density function
// at a distance dist
// (in radians).
func (n Normal) LogProb(dist float64) float64 {
	r := int(math.Round(dist / n.step))
	if r >= len(n.logPDF) {
		return n.logPDF[len(n.logPDF)-1]
	}
	return n.logPDF[r]
}

// Pix returns the underlying pixelation
// of a normal distribution.
func (n Normal) Pix() *earth.Pixelation {
	return n.pix
}

// Prob returns the value of the probability density function
// for a pixel at a distance dist
// (in radians).
func (n Normal) Prob(dist float64) float64 {
	r := int(math.Round(dist / n.step))
	if r >= len(n.pdf) {
		return 0
	}
	return n.pdf[r]
}

// Ring returns the value of the probability density function
// for a ring at a distance dist
// (in radians).
func (n Normal) Ring(dist float64) float64 {
	r := int(math.Round(dist / n.step))
	if r >= len(n.ring) {
		return 0
	}
	return n.ring[r]
}
