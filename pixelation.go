// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package earth

import (
	"fmt"
	"math"
	"math/rand"
	"sync"

	"gonum.org/v1/gonum/spatial/r3"
)

// A Pixelation is an equal area isolatitude pixelation.
//
// In this pixelation the Earth is divided in rings,
// and each ring is divided in pixels
// taking into account the diameter of the ring.
//
// To reduce the alignment of pixels between rings
// at the 180° meridian,
// odd rings have an offset of the size of a half pixel.
type Pixelation struct {
	eq      int     // pixels at the equator
	rings   []int   // index of the first pixel of each ring
	dStep   float64 // size (in degrees at the equator) of a pixel
	pixels  []Pixel
	perRing []int // number of pixels in each ring

	// Index allows a quick retrieval of pixels
	// using an equirectangular projection
	mu    sync.RWMutex
	cols  int
	iStep float64
	index []int
}

// NewPixelation returns a new pixelation
// with a given number of pixels
// at the equatorial parallel.
func NewPixelation(eq int) *Pixelation {
	if eq%2 != 0 {
		eq++
	}
	rings := eq/2 + 1
	dStep := 360 / float64(eq)

	pix := &Pixelation{
		eq:      eq,
		rings:   make([]int, rings),
		dStep:   dStep,
		perRing: make([]int, rings),
	}

	// add pixels
	for r := 0; r < rings; r++ {
		pix.rings[r] = len(pix.pixels)
		pix.addPixels(r)
		pix.perRing[r] = len(pix.pixels) - pix.rings[r]
	}

	// The index has a resolution
	// 1o times greater than the pixelation
	pix.cols = pix.eq * 10
	pix.iStep = 360 / float64(pix.cols)
	pix.index = make([]int, pix.cols*pix.cols/2)
	for i := range pix.index {
		pix.index[i] = -1
	}

	return pix
}

// Equator returns the number of pixels
// at the equatorial parallel.
func (pix *Pixelation) Equator() int {
	return pix.eq
}

// FirstPix returns the first pixel of a ring.
func (pix *Pixelation) FirstPix(ring int) Pixel {
	return pix.pixels[pix.rings[ring]]
}

// FromVector returns a pixel
// from a 3D vector of a geographic point.
// It panics if the vector is not valid
// (i.e. its norm is different from 1).
func (pix *Pixelation) FromVector(v r3.Vec) Pixel {
	// Set the tolerance to the 5%
	// but this is the square of the norm.
	if n2 := r3.Norm2(v); n2 < 0.9025 || n2 > 1.1025 {
		msg := fmt.Sprintf("invalid vector norm: %.3f", math.Sqrt(n2))
		panic(msg)
	}

	rLat := math.Asin(v.Z)
	lat := ToDegree(rLat)

	rLon := math.Atan2(v.Y, v.X)
	lon := ToDegree(rLon)

	return pix.getPixel(lat, lon)
}

// ID returns a pixel
// by its ID.
func (pix *Pixelation) ID(id int) Pixel {
	return pix.pixels[id]
}

// Len returns the number of pixels in the pixelation.
func (pix *Pixelation) Len() int {
	return len(pix.pixels)
}

// Pixel returns a pixel
// from a latitude and longitude coordinate pair.
// It panics if the coordinates are not valid.
func (pix *Pixelation) Pixel(lat, lon float64) Pixel {
	if lat < -90 || lat > 90 {
		msg := fmt.Sprintf("invalid latitude value: %.3f", lat)
		panic(msg)
	}
	if lon < -180 || lon > 180 {
		msg := fmt.Sprintf("invalid longitude value: %.3f", lon)
		panic(msg)
	}

	return pix.getPixel(lat, lon)
}

// PixPerRing returns the number of pixels in a ring.
func (pix *Pixelation) PixPerRing(ring int) int {
	return pix.perRing[ring]
}

// Random returns a random pixel from the pixelation.
func (pix *Pixelation) Random() Pixel {
	id := rand.Intn(len(pix.pixels))
	return pix.pixels[id]
}

// RandInRing returns a random pixel at a given ring.
func (pix *Pixelation) RandInRing(ring int) Pixel {
	id := pix.rings[ring] + rand.Intn(pix.perRing[ring])
	return pix.pixels[id]
}

// RingLat returns the latitude of a ring.
func (pix *Pixelation) RingLat(ring int) float64 {
	px := pix.pixels[pix.rings[ring]]
	return px.point.lat
}

// Rings returns the number of rings in the pixelation.
func (pix *Pixelation) Rings() int {
	return len(pix.rings)
}

// Step returns the size of a pixel in degrees
// at equator
// or its latitude size.
func (pix *Pixelation) Step() float64 {
	return pix.dStep
}

// AddPixels adds pixels to a pixelation ring.
func (pix *Pixelation) addPixels(r int) {
	lat := 90 - float64(r)*pix.dStep
	rLat := ToRad(lat)
	rStep := ToRad(pix.dStep)

	diameter := 2 * math.Pi * math.Cos(rLat)
	num := math.Round(diameter / rStep)
	if num == 0 {
		num = 1
	}
	ringStep := 360 / num
	for i := 0; i < int(num); i++ {
		lon := float64(i)*ringStep - 180
		if r%2 == 1 {
			lon += ringStep / 2
		}
		pt := NewPoint(lat, lon)
		px := Pixel{
			id:    len(pix.pixels),
			ring:  r,
			point: pt,
		}
		pix.pixels = append(pix.pixels, px)
	}
}

// Closest returns the closest pixel of a point
// from a ring.
func (pix *Pixelation) closest(ring int, pt Point) int {
	if ring > 0 {
		ring--
	}

	id := -1
	min := float64(2)
	for _, px := range pix.pixels[pix.rings[ring]:] {
		if px.ring > ring+2 {
			break
		}
		c2 := Chord2(pt, px.point)
		if c2 < min {
			min = c2
			id = px.id
		}
	}
	return id
}

// GetPixel returns a pixel from a latitude longitude pair.
func (pix *Pixelation) getPixel(lat, lon float64) Pixel {
	pos := pix.indexPos(lat, lon)

	pix.mu.RLock()
	id := pix.index[pos]
	pix.mu.RUnlock()

	if id != -1 {
		return pix.pixels[id]
	}

	pt := NewPoint(lat, lon)
	ring := int(math.Round((90 - lat) / pix.dStep))
	id = pix.closest(ring, pt)

	pix.mu.Lock()
	pix.index[pos] = id
	pix.mu.Unlock()

	return pix.pixels[id]
}

// IndexPos returns the position of a coordinate pair
// in an index.
func (pix *Pixelation) indexPos(lat, lon float64) int {
	x := int((lon + 180) / pix.iStep)
	if x == pix.cols {
		// points at 180 longitude
		// will set as -180 longitude
		x = 0
	}

	y := int((90 - lat) / pix.iStep)
	if y == pix.cols/2 {
		// points at -90 latitude
		// set to be a bit less than -90
		y = pix.cols/2 - 1
	}
	return y*pix.cols + x
}

// A Pixel is a pixel in a pixelation.
type Pixel struct {
	id    int
	ring  int
	point Point
}

// ID returns the index used to identify
// a pixel in a pixelation.
func (px Pixel) ID() int {
	return px.id
}

// Point returns the geographic point
// associate with a pixel.
func (px Pixel) Point() Point {
	return px.point
}

// Ring returns the ring of the pixel.
func (px Pixel) Ring() int {
	return px.ring
}
