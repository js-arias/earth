// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package rawdist_test

import (
	"math"
	"testing"

	"github.com/js-arias/earth"
	"github.com/js-arias/earth/stat/dist"
	"github.com/js-arias/earth/stat/rawdist"
)

func TestNormal(t *testing.T) {
	pix := earth.NewPixelation(360)
	lambda := 1.0
	n := dist.NewNormal(lambda, pix)

	u := rawdist.NewNormal(lambda, pix.Rings()-1)
	if got := u.Lambda(); got != lambda {
		t.Errorf("lambda: got %.6f, want %.6f\n", got, lambda)
	}

	np := earth.NorthPole

	// Test logProb function
	pdf := make([]float64, pix.Len())
	var sum float64
	for px := range pdf {
		pt := pix.ID(px).Point()
		p := math.Exp(u.LogProb(earth.Distance(np, pt)))
		pdf[px] = p
		sum += p
	}
	testProbValues(t, "logProb", pix, np, pdf, sum, n)

	// Test LogProbRingDist function
	pdf = make([]float64, pix.Len())
	sum = 0
	for px := range pdf {
		pt := pix.ID(px).Point()
		dist := earth.Distance(np, pt)
		d := math.Round(dist * float64(pix.Rings()-1) / math.Pi)
		p := math.Exp(u.LogProbRingDist(int(d)))
		pdf[px] = p
		sum += p
	}
	testProbValues(t, "logProbRingDist", pix, np, pdf, sum, n)

	// Test Prob function
	pdf = make([]float64, pix.Len())
	sum = 0
	for px := range pdf {
		pt := pix.ID(px).Point()
		p := u.Prob(earth.Distance(np, pt))
		pdf[px] = p
		sum += p
	}
	testProbValues(t, "Prob", pix, np, pdf, sum, n)

	// Test ProbRingDist function
	pdf = make([]float64, pix.Len())
	sum = 0
	for px := range pdf {
		pt := pix.ID(px).Point()
		dist := earth.Distance(np, pt)
		d := math.Round(dist * float64(pix.Rings()-1) / math.Pi)
		p := u.ProbRingDist(int(d))
		pdf[px] = p
		sum += p
	}
	testProbValues(t, "ProbRingDist", pix, np, pdf, sum, n)

	// Test scaled probabilities
	for px := 0; px < pix.Len(); px++ {
		pt := pix.ID(px).Point()
		dist := earth.Distance(np, pt)
		if u.Prob(dist) != u.ScaledProb(dist) {
			t.Errorf("scaled-prob: distance %.6f [pixel: %d], got %.6f, want %.6f\n", dist, px, u.ScaledProb(dist), u.Prob(dist))
		}
		d := math.Round(dist * float64(pix.Rings()-1) / math.Pi)
		if u.ScaledProbRingDist(int(d)) != u.ProbRingDist(int(d)) {
			t.Errorf("scaled-prob-ring-dist: distance %.6f [pixel: %d], got %.6f, want %.6f\n", dist, px, u.ScaledProbRingDist(int(d)), u.ProbRingDist(int(d)))
		}
	}
}

func testProbValues(t testing.TB, name string, pix *earth.Pixelation, rp earth.Point, rawVals []float64, sum float64, n dist.Normal) {
	t.Helper()

	for px, p := range rawVals {
		pt := pix.ID(px).Point()
		dist := earth.Distance(rp, pt)
		np := n.Prob(dist)

		// very small values can not be compared successfully
		if np < 1e-300 {
			continue
		}

		got := p / sum
		delta := math.Abs(got - np)
		if delta > 0.01 {
			t.Errorf("%s: pixel %d, distance %.6f, got %.6f [unscaled=%.6f], want %.6f [delta %.6f]", name, px, dist, got, p, np, delta)
		}
	}

}
