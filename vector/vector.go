// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package vector implements geological features
// using a vectorial data model.
package vector

import (
	"fmt"
	"strconv"
	"strings"
)

// Type is the type of a tectonic element.
type Type string

// Valid types of tectonic elements.
// This types are taken from the definitions
// of the GPML format
// (<https://www.gplates.org/docs/gpgim/>).
const (
	Basin     Type = "basin"
	Boundary  Type = "plate boundary"
	Coastline Type = "coastline"

	// A polygon that represents a boundary
	// between continental and oceanic crust.
	Continent Type = "continental boundary"

	// A large portion of a continental plate
	// that has been relatively undisturbed
	// since the Precambrian era.
	Craton Type = "craton"

	Fragment Type = "continental fragment"

	// A generic or unclassified feature.
	Generic Type = "generic"

	// A present day surface expression of a mantle plume.
	HotSpot Type = "hotspot"

	// A volcanic arc that is formed from magma rising
	// from a subduction oceanic plate.
	IslandArc Type = "island arc"

	// An extensive region of basalts
	// resulting from flood basalt volcanism.
	LIP Type = "large igneous province"

	// A part of a plate boundary
	// that no longer exists.
	PaleoBoundary Type = "paleo-boundary"

	// A passive continental boundary
	// indicating the change between continental and oceanic crust.
	Passive Type = "passive continental boundary"

	// A large-scale structural feature
	// associated with continental collision.
	Suture Type = "suture"

	// A crust fragment formed on a tectonic plate
	// and accreted to crust lying on another plate.
	Terrane Type = "terrane"
)

// A Feature is a tectonic feature.
type Feature struct {
	Name  string
	Type  Type
	Plate int // Plate ID

	// Temporal range of the feature
	// in years.
	Begin int64
	End   int64

	// Geographic coordinates of the feature
	Point   *Point
	Polygon Polygon
}

// A Point is a geographic point.
type Point struct {
	Lat float64
	Lon float64
}

// ParsePoint returns a point from an string pair
// that contains the latitude and longitude
// of a geographic point.
func ParsePoint(sLat, sLon string) (Point, error) {
	lat, err := strconv.ParseFloat(sLat, 64)
	if err != nil {
		return Point{360, 360}, fmt.Errorf("bad latitude value %q: %v", sLat, err)
	}
	if lat < -90 || lat > 90 {
		return Point{360, 360}, fmt.Errorf("bad latitude value %q", sLat)
	}

	lon, err := strconv.ParseFloat(sLon, 64)
	if err != nil {
		return Point{360, 360}, fmt.Errorf("bad longitude value %q: %v", sLon, err)
	}
	if lon < -180 || lon > 180 {
		return Point{360, 360}, fmt.Errorf("bad longitude value %q", sLon)
	}

	return Point{Lat: lat, Lon: lon}, nil
}

// A Polygon is an ordered collection of points
// that encloses an area.
type Polygon []Point

// ParsePolygon returns a polygon
// from a string that contains a list of coordinates
// (latitude and longitude)
// separated by spaces.
//
// For example:
//
//	85.151499473643639 180
//	83.701321128178307 180
//	83.870575085724681 178.17567629054798
//	85.08306008527849 178.86696019920456
//	85.151499473643639 180
func ParsePolygon(points string) (Polygon, error) {
	coord := strings.Fields(points)
	if len(coord)%2 != 0 {
		return nil, fmt.Errorf("invalid number of coordinates: %d", len(coord))
	}

	poly := make(Polygon, 0, len(coord)/2)
	for i := 0; i < len(coord); i += 2 {
		p, err := ParsePoint(coord[i], coord[i+1])
		if err != nil {
			return nil, err
		}
		poly = append(poly, p)
	}

	return poly, nil
}

// Bounds return the north and south coordinate
// defined for a polygon.
func (poly Polygon) bounds() (north, south float64) {
	north = -90
	south = 90

	for _, p := range poly {
		if p.Lat > north {
			north = p.Lat
		}
		if p.Lat < south {
			south = p.Lat
		}
	}
	return north, south
}
