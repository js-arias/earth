// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package earth_test

import (
	"math"
	"sync"
	"testing"

	"github.com/js-arias/earth"
)

func TestNewPixelation(t *testing.T) {
	eq := 36
	r := float64(eq) / (2 * math.Pi)

	// expected number of pixels
	// using the sphere area
	want := 4 * math.Pi * r * r

	pix := earth.NewPixelation(eq)
	got := float64(pix.Len())

	diff := got - want
	if diff < 0 {
		diff = -diff
	}
	if diff/want > 0.05 {
		t.Errorf("got %d pixels, want %.2f (error = %.2f%%)", int(got), want, diff)
	}

	rings := eq/2 + 1
	if pix.Rings() != rings {
		t.Errorf("got %d rings, want %d", pix.Rings(), rings)
	}
}

func TestPixelationPixel(t *testing.T) {
	tests := map[string]struct {
		lat, lon float64
		id       int
		ring     int
	}{
		"Tucumán":    {lat: -26, lon: -65, id: 29611, ring: 116},
		"North pole": {lat: 90, lon: 180},
		"South pole": {lat: -90, lon: -180, id: 41257, ring: 180},
		"Quito":      {lat: 0, lon: -78, id: 20551, ring: 90},
		"London":     {lat: 51, lon: 0, id: 4597, ring: 39},
		"Tokyo":      {lat: 35, lon: 139, id: 8912, ring: 55},
		"Anchorage":  {lat: 61, lon: -149, id: 2514, ring: 29},
	}

	eq := 360
	pix := earth.NewPixelation(eq)
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			pixHelper(t, pix, test.lat, test.lon, test.id, test.ring)
		})
	}
}

func TestPixelationRandom(t *testing.T) {
	eq := 360
	pix := earth.NewPixelation(eq)
	for i := 0; i < 100_000; i++ {
		px := pix.Random()
		pixHelper(t, pix, px.Point().Latitude(), px.Point().Longitude(), px.ID(), px.Ring())
	}
}

func pixHelper(t testing.TB, pix *earth.Pixelation, lat, lon float64, id, ring int) {
	t.Helper()

	pt := earth.NewPoint(lat, lon)
	px := pix.Pixel(lat, lon)

	dist := earth.Distance(pt, px.Point())
	if dist > 0.1 {
		t.Errorf("distance (from coord[lat=%.6f,lon=%.6f]): got %.6f", lat, lon, dist)
	}

	if got := px.ID(); got != id {
		t.Errorf("ID (from coord[lat=%.6f,lon=%.6f]): got %d, want %d", lat, lon, got, id)
	}

	if got := px.Ring(); got != ring {
		t.Errorf("ring (from coord[lat=%.6f,lon=%.6f]), got %d, want %d", lat, lon, got, ring)
	}

	np := pix.ID(id)
	dist = earth.Distance(pt, np.Point())
	if dist > 0.1 {
		t.Errorf("distance (from ID %d): %.6f", id, dist)
	}
}

func TestPixelationFromVector(t *testing.T) {
	tests := map[string]struct {
		lat, lon float64
	}{
		"Tucumán":    {lat: -26, lon: -65},
		"North pole": {lat: 90, lon: 180},
		"South pole": {lat: -90, lon: -180},
		"Quito":      {lat: 0, lon: -78},
		"London":     {lat: 51, lon: 0},
		"Tokyo":      {lat: 35, lon: 139},
		"Anchorage":  {lat: 61, lon: -149},
	}

	eq := 360
	pix := earth.NewPixelation(eq)

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			pt := earth.NewPoint(test.lat, test.lon)
			px := pix.Pixel(test.lat, test.lon)
			vecHelper(t, pix, pt, px.ID())
		})
	}
}

func TestPixelationFromVectorRandom(t *testing.T) {
	eq := 360
	pix := earth.NewPixelation(eq)
	for i := 0; i < 100_000; i++ {
		px := pix.Random()
		vecHelper(t, pix, px.Point(), px.ID())
	}
}

func vecHelper(t testing.TB, pix *earth.Pixelation, pt earth.Point, id int) {
	t.Helper()

	v := pt.Vector()
	px := pix.FromVector(v)

	dist := earth.Distance(pt, px.Point())
	if dist > 0.1 {
		t.Errorf("distance (from vector[x=%.6f,y=%.6f,z=%.6f]): got %.6f", v.X, v.Y, v.Z, dist)
	}

	if got := px.ID(); got != id {
		t.Errorf("ID (from vector[x=%.6f,y=%.6f,z=%.6f]): got %d, want %d", v.X, v.Y, v.Z, got, id)
	}
}

func TestPixelLocationRace(t *testing.T) {
	eq := 360
	pix := earth.NewPixelation(eq)

	var done sync.WaitGroup
	done.Add(2)
	go func() {
		for i := 0; i < 100_000; i++ {
			px := pix.Random()
			pixHelper(t, pix, px.Point().Latitude(), px.Point().Longitude(), px.ID(), px.Ring())
			t.Logf("pix: %d\n", i)
		}
		done.Done()
	}()

	go func() {
		for i := 0; i < 100_000; i++ {
			px := pix.Random()
			vecHelper(t, pix, px.Point(), px.ID())
			t.Logf("vec: %d\n", i)
		}
		done.Done()
	}()

	done.Wait()
}

func TestPixelRndDistance(t *testing.T) {
	eq := 360
	pix := earth.NewPixelation(eq)

	for i := 0; i < 1000; i++ {
		px := pix.Random()
		for ref := 0; ref < pix.Len(); ref++ {
			rp := pix.ID(ref)
			dist := earth.Distance(px.Point(), rp.Point())
			if math.IsNaN(dist) {
				t.Errorf("distance between pix %d and %d is NaN", px.ID(), rp.ID())
			}
		}
	}
}

func TestPixelationRings(t *testing.T) {
	tests := []struct {
		num   int
		first int
	}{
		{num: 1},
		{first: 1, num: 6},
		{first: 7, num: 12},
		{first: 19, num: 18},
		{first: 37, num: 23},
		{first: 60, num: 28},
		{first: 88, num: 31},
		{first: 119, num: 34},
		{first: 153, num: 35},
		{first: 188, num: 36},
		{first: 224, num: 35},
		{first: 259, num: 34},
		{first: 293, num: 31},
		{first: 324, num: 28},
		{first: 352, num: 23},
		{first: 375, num: 18},
		{first: 393, num: 12},
		{first: 405, num: 6},
		{first: 411, num: 1},
	}
	eq := 36
	pix := earth.NewPixelation(eq)

	for r, test := range tests {
		if px := pix.FirstPix(r); px.ID() != test.first {
			t.Errorf("first pixel at ring %d: got %d, want %d", r, px.ID(), test.first)
		}
		if ppr := pix.PixPerRing(r); ppr != test.num {
			t.Errorf("pixels at ring %d: got %d, want %d", r, ppr, test.num)
		}
	}
}
