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
// (in 1/radians^2),
// and gcd is the great circle distance.
type Normal struct {
	pix    *earth.Pixelation
	step   float64 // step of a ring in radians
	lambda float64 // concentration parameter

	maxRing float64 // maximum ring probability

	pdf    []float64
	cdf    []float64
	ring   []float64
	logPDF []float64
}

// NewNormal returns a discretized spherical normal,
// using lambda as the concentration parameter
// (in 1/radian^2 units)
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
	var maxRing float64
	pdf := make([]float64, rings)
	logSum := math.Log(sum)
	for i := range logPDF {
		r := ring[i] / sum
		ring[i] = r
		if r > maxRing {
			maxRing = r
		}

		cdf[i] = cdf[i] / sum
		logPDF[i] = logPDF[i] - logSum
		pdf[i] = math.Exp(logPDF[i])
	}

	return Normal{
		pix:     pix,
		step:    rStep,
		lambda:  lambda,
		maxRing: maxRing,
		pdf:     pdf,
		cdf:     cdf,
		ring:    ring,
		logPDF:  logPDF,
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
// (in 1/radians^2)
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

// LogProbRingDist returns the natural logarithm
// of the probability density function
// at a given ring distance
// i.e. the ring of a pixel,
// if one of the pixels is rotated to the north pole.
func (n Normal) LogProbRingDist(rDist int) float64 {
	return n.logPDF[rDist]
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

// ProbRingDist returns the the value of the probability density function
// at a given ring distance
// i.e. the ring of a pixel,
// if one of the pixels is rotated to the north pole.
func (n Normal) ProbRingDist(rDist int) float64 {
	return n.pdf[rDist]
}

// Rand returns a random pixel
// from the underlying pixelation
// draw from an spherical normal
// which mean is the pixel u.
func (n Normal) Rand(u earth.Pixel) earth.Pixel {
	uPt := u.Point()

	// For small lambda values
	// sample from an uniform distribution over the globe.
	// The limit was found empirically
	if n.lambda <= 8 {
		logP := n.logPDF[0]
		for {
			tp := n.pix.Random()
			dist := earth.Distance(uPt, tp.Point())

			logT := n.LogProb(dist)
			accept := math.Exp(logT - logP)
			if rand.Float64() < accept {
				return tp
			}
		}
	}

	// for large lambda values
	// sample using the radius
	for {
		dist := rand.Float64() * math.Pi
		p := n.Ring(dist) / n.maxRing
		if rand.Float64() >= p {
			continue
		}

		b := rand.Float64() * 2 * math.Pi
		pt := earth.Destination(uPt, dist, b)
		return n.pix.Pixel(pt.Latitude(), pt.Longitude())
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
