// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package dist provides spherical distribution types.
package dist

import (
	"math"
	"time"

	"github.com/js-arias/earth"
	"golang.org/x/exp/rand"
	"golang.org/x/exp/slices"
)

func init() {
	rand.Seed(uint64(time.Now().UnixNano()))
}

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

// QuantileChord2 returns the square of the Euclidean chord distance
// for the maximum distance
// that is inside the indicated cumulative density.
//
// This is useful because sometimes we want to know
// if a given pixel is inside or outside a critical CDF value
// and then using the great circle distance.
func (n Normal) QuantileChord2(cd float64) float64 {
	r, _ := slices.BinarySearch(n.cdf, cd)
	px := n.pix.FirstPix(r)
	np := n.pix.Pixel(90, 0)
	return earth.Chord2(px.Point(), np.Point())
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

// Rand returns a random pixel
// from the underlying pixelation
// draw from an spherical normal
// which mean is the pixel u.
//
// It use a simple rejection-sampling algorithm,
// with an uniform proposal distribution for each pixel.
// The proposal pixel is accepted with a probability:
//
//	p = SN(x) / (Uniform(x) * c)
//
// where SN is the PDF of a pixel x with the spherical normal,
// Uniform is the PDF of the uniform distribution
// and c is a constant chosen such that SN(x) < Uniform(x)*c for all x.
func (n Normal) Rand(u earth.Pixel) earth.Pixel {
	uPt := u.Point()
	logP := -math.Log(float64(n.pix.Len()))

	// As the maximum probability point for the normal
	// is mean pixel,
	// we guarantee that any value of SN(x)
	// is smaller than Uniform(x)*c.
	c := math.Exp(n.logPDF[0]-logP) * 2
	for {
		tp := n.pix.Random()
		dist := earth.Distance(uPt, tp.Point())

		logT := n.LogProb(dist)
		accept := math.Exp(logT-logP) / c
		if rand.Float64() < accept {
			return tp
		}
	}
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
