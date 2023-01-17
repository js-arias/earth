// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package dist_test

import (
	"math"
	"testing"

	"github.com/js-arias/earth"
	"github.com/js-arias/earth/dist"
)

func TestNormal(t *testing.T) {
	pix := earth.NewPixelation(360)
	lambda := 1.0
	n := dist.NewNormal(lambda, pix)

	var sum float64
	px := pix.Random()
	for i := 0; i < pix.Len(); i++ {
		np := pix.ID(i)
		dist := earth.Distance(px.Point(), np.Point())
		p := n.Prob(dist)
		sum += p

		// very small values can not be compared successfully
		if p < 1e-300 {
			continue
		}
		got := n.LogProb(dist)
		want := math.Log(p)
		delta := math.Abs(got - want)
		if delta > 0.01 {
			t.Errorf("logPDF: distance %.6f [prob %.6f], got %.6f, want %.6f [delta %.6f]", dist, p, got, want, delta)
		}
	}

	diff := math.Abs(1 - sum)
	if diff > 0.05 {
		t.Errorf("pdf: got %.6f sum, want %.6f (error = %.2f%%)", sum, 1.0, diff*100)
	}

	sum = 0
	np := pix.Pixel(90, 0)
	for i := 0; i < pix.Rings(); i++ {
		rp := pix.Pixel(90-float64(i), 0)
		dist := earth.Distance(np.Point(), rp.Point())
		sum += n.Ring(dist)

		got := n.CDF(dist)
		delta := math.Abs(got - sum)
		if delta > 0.01 {
			t.Errorf("CDF: distance %.6f, got %.6f, want %.6f", dist, got, sum)
		}
	}

	if n.Pix() != pix {
		t.Error("Pixelation: unable to retrieve source pixelation")
	}
	if n.Lambda() != lambda {
		t.Errorf("Lambda: got %.6f, want %.6f", n.Lambda(), lambda)
	}
}

func TestNormalMean(t *testing.T) {
	pix := earth.NewPixelation(360)

	tests := map[string]struct {
		p1 int
		p2 int
		m  int
	}{
		"south polar circle": {
			p1: pix.Pixel(90, 0).ID(),
			p2: pix.Pixel(-66, 0).ID(),
			m:  16329,
		},
		"equator": {
			p1: pix.Pixel(90, 0).ID(),
			p2: pix.Pixel(0, 0).ID(),
			m:  6037,
		},
	}

	n := dist.NewNormal(1, pix)
	for name, test := range tests {
		p1 := pix.ID(test.p1)
		p2 := pix.ID(test.p2)

		max := -math.MaxFloat64
		var b int
		for i := 0; i < pix.Len(); i++ {
			tp := pix.ID(i)
			dist1 := earth.Distance(p1.Point(), tp.Point())
			dist2 := earth.Distance(p2.Point(), tp.Point())
			p := n.LogProb(dist1) + n.LogProb(dist2)
			if p > max {
				max = p
				b = i
			}
		}

		if b != test.m {
			t.Errorf("%s: mean pixel: got %d, want %d", name, b, test.m)
		}
	}

	// Antipodes
	// in this case the mean should be any point at equator
	np := pix.Pixel(90, 0)
	sp := pix.Pixel(-90, 0)
	max := -math.MaxFloat64
	lat := 360.0
	for i := 0; i < pix.Len(); i++ {
		tp := pix.ID(i)
		dist1 := earth.Distance(np.Point(), tp.Point())
		dist2 := earth.Distance(sp.Point(), tp.Point())
		p := n.LogProb(dist1) + n.LogProb(dist2)
		if p > max {
			max = p
			lat = tp.Point().Latitude()
		}
	}

	if lat != 0 {
		t.Errorf("poles: mean: got %.6f, want %.6f", lat, 0.0)
	}
}

func TestInvChord2(t *testing.T) {
	pix := earth.NewPixelation(360)
	n := dist.NewNormal(1, pix)
	bound := 0.95

	np := pix.Pixel(90, 0)

	c2 := n.InvChord2(bound)
	for i := 0; i < pix.Len(); i++ {
		px := pix.ID(i)
		ch2 := earth.Chord2(np.Point(), px.Point())
		dist := earth.Distance(np.Point(), px.Point())
		if ch2 >= c2 {
			if n.CDF(dist) <= bound {
				t.Errorf("%.6f: inside: at chord distance %.6f [%.6f], CDF is %.6f", c2, ch2, dist, n.CDF(dist))
			}
			continue
		}
		if n.CDF(dist) > bound {
			t.Errorf("%.6f: outside: at chord distance %.6f [%.6f], CDF is %.6f", c2, ch2, dist, n.CDF(dist))
		}
	}
}
