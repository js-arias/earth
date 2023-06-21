// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package earth_test

import (
	"math"
	"testing"

	"github.com/js-arias/earth"
)

func TestDistMat(t *testing.T) {
	pix := earth.NewPixelation(360)
	m, err := earth.NewDistMat(pix)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < 10000; i++ {
		px1 := pix.Random()
		px2 := pix.Random()

		d := earth.Distance(px1.Point(), px2.Point())
		rd := int(math.Round(d / earth.ToRad(pix.Step())))

		got := m.At(px1.ID(), px2.ID())
		if got != rd {
			t.Errorf("pixels %d, %d: ring distance %d, want %d", px1.ID(), px2.ID(), got, rd)
		}
	}
}
