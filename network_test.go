// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package earth_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/js-arias/earth"
)

func TestNewNetwork(t *testing.T) {
	eq := 120
	pix := earth.NewPixelation(eq)
	r := earth.ToRad(360 / float64(eq))
	r *= math.Sqrt2

	want := make([][]int, pix.Len())
	for px1 := range pix.Len() {
		pt1 := pix.ID(px1).Point()
		var n []int
		for px2 := range pix.Len() {
			if px1 == px2 {
				n = append(n, px2)
				continue
			}
			pt2 := pix.ID(px2).Point()
			if earth.Distance(pt1, pt2) < r {
				n = append(n, px2)
			}
		}
		want[px1] = n
	}

	net := earth.NewNetwork(pix)
	for px, n := range net {
		if !reflect.DeepEqual(n, want[px]) {
			t.Errorf("pixel %d: got %v, want %v", px, n, want[px])
		}
	}
}

