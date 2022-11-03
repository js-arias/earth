// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package earth implements an spherical model of the Earth.
package earth

import (
	"math"

	"gonum.org/v1/gonum/spatial/r3"
)

// ToDegree transform a radian angle into degrees.
func ToDegree(angle float64) float64 {
	return angle * 180 / math.Pi
}

// ToRad transform a degree angle into radians.
func ToRad(angle float64) float64 {
	return angle * math.Pi / 180
}

// A Point is a geographic point
// on the surface of the unit length sphere.
type Point struct {
	lat, lon float64
	vec      r3.Vec
}

// NewPoint returns a geographic point
// from a pair of geographic coordinates.
func NewPoint(lat, lon float64) Point {
	rLat := ToRad(lat)
	rLon := ToRad(lon)
	return Point{
		lat: rLat,
		lon: rLon,
		vec: r3.Vec{
			X: math.Cos(rLat) * math.Cos(rLon),
			Y: math.Cos(rLat) * math.Sin(rLon),
			Z: math.Sin(rLat),
		},
	}
}

// Latitude returns the latitude of a point.
func (p Point) Latitude() float64 {
	return p.lat
}

// Longitude returns the longitude of a point.
func (p Point) Longitude() float64 {
	return p.lon
}

// Vector returns the 2D vector representation of a point.
func (p Point) Vector() r3.Vec {
	return p.vec
}

// Earth poles
var NorthPole = NewPoint(90, 0)
var SouthPole = NewPoint(-90, 0)

// Distance returns the great circle distance,
// in radians,
// between two geographic points.
func Distance(p, q Point) float64 {
	dot := r3.Dot(p.vec, q.vec)
	if dot > 1 {
		dot = 1
	}
	if dot < -1 {
		dot = -1
	}
	return math.Acos(dot)
}
