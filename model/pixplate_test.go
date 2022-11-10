// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package model_test

import (
	"reflect"
	"testing"

	"github.com/js-arias/earth"
	"github.com/js-arias/earth/model"
)

func TestNewPixPlate(t *testing.T) {
	pp := makePixPlate(t)
	testPixPlate(t, pp)
}

func makePixPlate(t testing.TB) *model.PixPlate {
	t.Helper()

	pp := model.NewPixPlate(earth.NewPixelation(360))

	data := []struct {
		name       string
		lat, lon   float64
		begin, end int64
		plate      int
	}{
		{
			name:  "Parana",
			lat:   -26,
			lon:   -65,
			begin: 600_000_000,
			plate: 202,
		},
		{
			name:  "Antarctica",
			lat:   -90,
			lon:   -180,
			begin: 600_000_000,
			plate: 802,
		},
	}
	for _, d := range data {
		pp.Add(d.plate, d.name, d.lat, d.lon, d.begin, d.end)
	}

	square := []int{17051, 17052, 17053, 17054, 17055, 17406, 17407, 17408, 17409, 17410, 17763, 17764, 17765, 17766, 17767, 18119, 18120, 18121, 18122, 18123, 18477, 18478, 18479, 18480, 18481}
	pp.AddPixels(59_999, "square", square, 140_000_000, 20_000_000)

	return pp
}

func testPixPlate(t testing.TB, pp *model.PixPlate) {
	t.Helper()

	if eq := pp.Pixelation().Equator(); eq != 360 {
		t.Errorf("pixelation: got %d pixels at the equator, want %d", eq, 360)
	}

	plates := []int{202, 802, 59_999}
	if p := pp.Plates(); !reflect.DeepEqual(p, plates) {
		t.Errorf("plates: got %v, want %v", p, plates)
	}

	tests := map[string]struct {
		plate int
		pix   []int
		begin int64
		end   int64
	}{
		"Parana": {
			plate: 202,
			pix:   []int{29611},
			begin: 600_000_000,
		},
		"Antarctica": {
			plate: 802,
			pix:   []int{41257},
			begin: 600_000_000,
		},
		"square": {
			plate: 59999,
			pix:   []int{17051, 17052, 17053, 17054, 17055, 17406, 17407, 17408, 17409, 17410, 17763, 17764, 17765, 17766, 17767, 18119, 18120, 18121, 18122, 18123, 18477, 18478, 18479, 18480, 18481},
			begin: 140_000_000,
			end:   20_000_000,
		},
	}
	for name, test := range tests {
		pix := pp.Pixels(test.plate)
		if !reflect.DeepEqual(pix, test.pix) {
			t.Errorf("%s-pixels: got %v, want %v", name, pix, test.pix)
		}

		for _, id := range pix {
			w := model.PixAge{
				Name:  name,
				ID:    id,
				Plate: test.plate,
				Begin: test.begin,
				End:   test.end,
			}
			px := pp.Pixel(test.plate, id)
			if !reflect.DeepEqual(px, w) {
				t.Errorf("%s-pixel %d: got %v, want %v", name, id, px, w)
			}
		}
	}
}
