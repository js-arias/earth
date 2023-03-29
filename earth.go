// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package earth implements an spherical model of the Earth.
package earth

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/spatial/r3"
)

const (
	// Arithmetic mean radius of Earth in meters
	// after Moritz (1980) Geodetic Reference System 1980
	// Resolution 1 at the XVII General Assembly of the IUGG.
	Radius = 6_371_008

	// Age of Earth in years.
	Age = 4_540_000_000
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
// It panics if the coordinates are not valid.
func NewPoint(lat, lon float64) Point {
	if lat < -90 || lat > 90 {
		msg := fmt.Sprintf("invalid latitude value: %.3f", lat)
		panic(msg)
	}
	if lon < -180 || lon > 180 {
		msg := fmt.Sprintf("invalid longitude value: %.3f", lon)
		panic(msg)
	}

	rLat := ToRad(lat)
	rLon := ToRad(lon)
	return Point{
		lat: lat,
		lon: lon,
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

// Chord2 returns the square of the Euclidean chord distance.
func Chord2(p, q Point) float64 {
	v := r3.Sub(p.vec, q.vec)
	return r3.Norm2(v)
}

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

// Bearing returns the direction angle
// between a meridian and the great circle line
// that connect two points,
// from the point p.
// The resulting angle is in radians,
// 0 being north,
// pi/2 east,
// pi south,
// and 3pi/2 west,
func Bearing(p, q Point) float64 {
	pLat := ToRad(p.lat)
	qLat := ToRad(q.lat)
	dLon := ToRad(q.lon) - ToRad(p.lon)
	x := math.Cos(qLat) * math.Sin(dLon)
	y := math.Cos(pLat)*math.Sin(qLat) - math.Sin(pLat)*math.Cos(qLat)*math.Cos(dLon)

	b := math.Atan2(x, y)
	if b < 0 {
		b = 2*math.Pi + b
	}
	return b
}

// Destination returns the destination point
// of a trip starting at point p,
// given a bearing and a distance
// (in radians).
func Destination(p Point, dist, bearing float64) Point {
	pLat := ToRad(p.lat)

	sinLat := math.Sin(pLat)*math.Cos(dist) + math.Cos(pLat)*math.Sin(dist)*math.Cos(bearing)
	rLat := math.Asin(sinLat)
	tanLonX := math.Sin(bearing) * math.Sin(dist) * math.Cos(pLat)
	tanLonY := math.Cos(dist) - math.Sin(pLat)*math.Sin(rLat)
	lon := p.lon + ToDegree(math.Atan2(tanLonX, tanLonY))
	if lon > 180 {
		lon = lon - 360
	}
	if lon < -180 {
		lon = 360 + lon
	}

	return NewPoint(ToDegree(rLat), lon)
}
