// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package earth_test

import (
	"math"
	"testing"

	"github.com/js-arias/earth"
)

func TestPointDistance(t *testing.T) {
	tests := map[string]struct {
		p1, p2 earth.Point
		dist   float64
	}{
		"Cape Town - Stockholm": {
			p1:   earth.NewPoint(-34, 18),
			p2:   earth.NewPoint(59, 18),
			dist: earth.ToRad(93),
		},
		"Cox & Hart, box 3.2 (left)": {
			p1:   earth.NewPoint(30, 40),
			p2:   earth.NewPoint(-30, 110),
			dist: earth.ToRad(90),
		},
		"Cox & Hart, box 3.2 (right)": {
			p1:   earth.NewPoint(60, -120),
			p2:   earth.NewPoint(-70, 120),
			dist: earth.ToRad(150),
		},
		"Antipodes": {
			p1:   earth.NewPoint(30, 30),
			p2:   earth.NewPoint(-30, -150),
			dist: earth.ToRad(180),
		},
		"Close": {
			p1:   earth.NewPoint(0, 20),
			p2:   earth.NewPoint(0, 21),
			dist: earth.ToRad(1),
		},
		"equal": {
			p1: earth.NewPoint(-44, 146),
			p2: earth.NewPoint(-44, 146),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := earth.Distance(test.p1, test.p2)
			if math.IsNaN(got) {
				t.Errorf("%s: NaN distance, want %.6f", name, test.dist)
			}
			diff := got - test.dist
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.1 {
				t.Errorf("%s: got %.6f, want %.6f (error = %.6f rad)", name, got, test.dist, diff)
			}
		})
	}
}

func TestBearing(t *testing.T) {
	tests := map[string]struct {
		p1, p2  earth.Point
		bearing float64
	}{
		"Kansas - St. Louis": {
			p1:      earth.NewPoint(39.099912, -94.581213),
			p2:      earth.NewPoint(38.627089, -90.200203),
			bearing: 1.684463,
		},
		"Tasmania - Tucuman": {
			p1:      earth.NewPoint(-42, 147),
			p2:      earth.NewPoint(-26, -65),
			bearing: 2.623630,
		},
		"Tasmania - Cairo": {
			p1:      earth.NewPoint(-42, 147),
			p2:      earth.NewPoint(30, 31),
			bearing: 4.862267,
		},
		"Tasmania - Los Angeles": {
			p1:      earth.NewPoint(-42, 147),
			p2:      earth.NewPoint(34, -118),
			bearing: 1.152416,
		},
		"Tasmania - Beijing": {
			p1:      earth.NewPoint(-42, 147),
			p2:      earth.NewPoint(39, 116),
			bearing: 5.870185,
		},
		"Tasmania - Maputo": {
			p1:      earth.NewPoint(-42, 147),
			p2:      earth.NewPoint(-25, 32),
			bearing: 4.105445,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := earth.Bearing(test.p1, test.p2)
			diff := got - test.bearing
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.1 {
				t.Errorf("%s: got %.6f, want %.6f (error = %.6f rad)", name, got, test.bearing, diff)
			}
		})
	}
}

func TestDestination(t *testing.T) {
	tests := map[string]struct {
		p1, p2 earth.Point
	}{
		"Kansas - St. Louis": {
			p1: earth.NewPoint(39.099912, -94.581213),
			p2: earth.NewPoint(38.627089, -90.200203),
		},
		"Tasmania - Tucuman": {
			p1: earth.NewPoint(-42, 147),
			p2: earth.NewPoint(-26, -65),
		},
		"Tasmania - Cairo": {
			p1: earth.NewPoint(-42, 147),
			p2: earth.NewPoint(30, 31),
		},
		"Tasmania - Los Angeles": {
			p1: earth.NewPoint(-42, 147),
			p2: earth.NewPoint(34, -118),
		},
		"Tasmania - Beijing": {
			p1: earth.NewPoint(-42, 147),
			p2: earth.NewPoint(39, 116),
		},
		"Tasmania - Maputo": {
			p1: earth.NewPoint(-42, 147),
			p2: earth.NewPoint(-25, 32),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			dist := earth.Distance(test.p1, test.p2)
			b := earth.Bearing(test.p1, test.p2)

			got := earth.Destination(test.p1, dist, b)

			d := earth.Distance(test.p2, got)
			if math.IsNaN(d) {
				t.Errorf("%s: NaN distance, want %.6f", name, 0.0)
			}
			if d > 0.01 {
				t.Errorf("%s: got %v, want %v [distance = %.6f]", name, got, test.p2, d)
			}
		})
	}

}
