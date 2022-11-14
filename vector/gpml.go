// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package vector

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/js-arias/earth"
)

// DecodeGPML returns an slice of vector features
// from a GPML encoded file.
//
// The GPML format is an implementation of XML
// for tectonic plates modelling,
// and is the main format used by [GPlates] software.
// For a formal description of the GPML format
// see [GPlates GPML documentation].
//
// [GPlates]: https://www.gplates.org
// [GPlates GPML documentation]: https://www.gplates.org/docs/gpgim/
func DecodeGPML(r io.Reader) ([]Feature, error) {
	d := xml.NewDecoder(r)
	c := collection{}
	if err := d.Decode(&c); err != nil {
		return nil, fmt.Errorf("unable to decode GPML: %v", err)
	}

	coll := c.features()
	fs := make([]Feature, 0, len(coll))
	for _, cf := range coll {
		begin, err := cf.begin()
		if err != nil {
			return nil, fmt.Errorf("feature %s [plate %d]: %v", cf.Name, cf.Plate, err)
		}
		end, err := cf.end()
		if err != nil {
			return nil, fmt.Errorf("feature %s [plate %d]: %v", cf.Name, cf.Plate, err)
		}

		pp, err := cf.polygons()
		if err != nil {
			return nil, fmt.Errorf("feature %s [plate %d]: %v", cf.Name, cf.Plate, err)
		}

		for _, p := range pp {
			f := Feature{
				Name:    cf.Name,
				Type:    cf.tp,
				Plate:   cf.Plate,
				Begin:   begin,
				End:     end,
				Polygon: p,
			}

			fs = append(fs, f)
		}
		if cf.Point != "" {
			coord := strings.Fields(cf.Point)
			if len(coord) != 2 {
				return nil, fmt.Errorf("feature %s [plate %d]: bad point: %s", cf.Name, cf.Plate, cf.Point)
			}
			pt, err := ParsePoint(coord[0], coord[1])
			if err != nil {
				return nil, fmt.Errorf("feature %s [plate %d]: bad point: %v", cf.Name, cf.Plate, err)
			}
			f := Feature{
				Name:  cf.Name,
				Type:  cf.tp,
				Plate: cf.Plate,
				Begin: begin,
				End:   end,
				Point: &pt,
			}

			fs = append(fs, f)
		}
	}
	return fs, nil
}

// MillionYears is used to transform GPML ages
// (a float in million years)
// to an integer in years.
const millionYears = 1_000_000

// A collection is a collection of geological features.
type collection struct {
	XMLName xml.Name `xml:"FeatureCollection"`

	// Features
	Basin         []feature `xml:"featureMember>Basin"`
	Boundary      []feature `xml:"featureMember>TopologicalClosedPlateBoundary"`
	Coastline     []feature `xml:"featureMember>Coastline"`
	Continent     []feature `xml:"featureMember>ClosedContinentalBoundary"`
	Craton        []feature `xml:"featureMember>Craton"`
	Fragment      []feature `xml:"featureMember>ContinentalFragment"`
	Generic       []feature `xml:"featureMember>UnclassifiedFeature"`
	HotSpot       []feature `xml:"featureMember>HotSpot"`
	IslandArc     []feature `xml:"featureMember>IslandArc"`
	LIP           []feature `xml:"featureMember>LargeIgneousProvince"`
	PaleoBoundary []feature `xml:"featureMember>InferredPaleoBoundary"`
	Passive       []feature `xml:"featureMember>PassiveContinentalBoundary"`
	Suture        []feature `xml:"featureMember>Suture"`
	Terrane       []feature `xml:"featureMember>TerraneBoundary"`
}

func (c collection) features() []feature {
	var f []feature
	for _, v := range c.Basin {
		v.tp = Basin
		f = append(f, v)
	}
	for _, v := range c.Boundary {
		v.tp = Boundary
		f = append(f, v)
	}
	for _, v := range c.Coastline {
		v.tp = Coastline
		f = append(f, v)
	}
	for _, v := range c.Continent {
		v.tp = Continent
		f = append(f, v)
	}
	for _, v := range c.Craton {
		v.tp = Craton
		f = append(f, v)
	}
	for _, v := range c.Fragment {
		v.tp = Fragment
		f = append(f, v)
	}
	for _, v := range c.Generic {
		v.tp = Generic
		f = append(f, v)
	}
	for _, v := range c.HotSpot {
		v.tp = HotSpot
		f = append(f, v)
	}
	for _, v := range c.IslandArc {
		v.tp = IslandArc
		f = append(f, v)
	}
	for _, v := range c.LIP {
		v.tp = LIP
		f = append(f, v)
	}
	for _, v := range c.PaleoBoundary {
		v.tp = PaleoBoundary
		f = append(f, v)
	}
	for _, v := range c.Passive {
		v.tp = Passive
		f = append(f, v)
	}
	for _, v := range c.Suture {
		v.tp = Suture
		f = append(f, v)
	}
	for _, v := range c.Terrane {
		v.tp = Terrane
		f = append(f, v)
	}

	return f
}

// A feature is a geographic polygon,
// a boundary,
// or a point,
// associated with a tectonic plate.
type feature struct {
	Name   string `xml:"name"`
	Plate  int    `xml:"reconstructionPlateId>ConstantValue>value"`
	Period period `xml:"validTime>TimePeriod"`

	Point    string    `xml:"position>Point>pos"`
	Boundary []polygon `xml:"boundary>ConstantValue>value>Polygon"`
	Outline  []polygon `xml:"outlineOf>ConstantValue>value>Polygon"`
	Line     []polygon `xml:"centerLineOf>ConstantValue>value>Polygon"`
	Generic  []polygon `xml:"unclassifiedGeometry>ConstantValue>value>Polygon"`

	tp Type
}

// Begin returns the maximum (oldest) age,
// in years,
// of a feature.
func (f feature) begin() (int64, error) {
	if f.Period.Begin == "http://gplates.org/times/distantPast" {
		return earth.Age, nil
	}
	age, err := strconv.ParseFloat(f.Period.Begin, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid 'begin' time: %v", err)
	}
	return int64(age * millionYears), nil
}

// End returns the minimum (youngest) age,
// in years,
// of a feature.
func (f feature) end() (int64, error) {
	if f.Period.End == "http://gplates.org/times/distantFuture" {
		return 0, nil
	}
	age, err := strconv.ParseFloat(f.Period.End, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid 'end' time: %v", err)
	}
	return int64(age * millionYears), nil
}

// Polygons returns the polygons
// of a feature.
func (f feature) polygons() ([]Polygon, error) {
	var pp []Polygon

	bp, err := parsePolygons(f.Boundary)
	if err != nil {
		return nil, fmt.Errorf("boundary polygon: %v", err)
	}
	pp = append(pp, bp...)

	gp, err := parsePolygons(f.Generic)
	if err != nil {
		return nil, fmt.Errorf("generic polygon: %v", err)
	}
	pp = append(pp, gp...)

	ln, err := parsePolygons(f.Line)
	if err != nil {
		return nil, fmt.Errorf("line polygon: %v", err)
	}
	pp = append(pp, ln...)

	ol, err := parsePolygons(f.Outline)
	if err != nil {
		return nil, fmt.Errorf("outline polygon: %v", err)
	}
	pp = append(pp, ol...)

	if len(pp) == 0 {
		return nil, nil
	}
	return pp, nil
}

// A period is the time period
// for a geological feature.
type period struct {
	Begin string `xml:"begin>TimeInstant>timePosition"`
	End   string `xml:"end>TimeInstant>timePosition"`
}

// A polygon is a collection of points
// enclosing a surface.
type polygon struct {
	PosList string `xml:"exterior>LinearRing>posList"`
}

func parsePolygons(ps []polygon) ([]Polygon, error) {
	var pp []Polygon

	for _, p := range ps {
		np, err := ParsePolygon(p.PosList)
		if err != nil {
			return nil, err
		}
		if len(np) == 0 {
			continue
		}
		pp = append(pp, np)
	}
	if len(pp) == 0 {
		return nil, nil
	}
	return pp, nil
}
