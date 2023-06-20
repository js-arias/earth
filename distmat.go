// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package earth

import (
	"fmt"
	"math"
)

// DistMat is a distance matrix for the pixels
// in a pixelation.
// It store not the real distance,
// but the "ring" distance,
// i.e. the ring of a pixel,
// if one of the pixels is rotated to the north pole.
type DistMat struct {
	pix  *Pixelation
	rows int
	m    []uint8
}

// NewDistMat creates a new distance matrix
// from the indicated pixelation.
// To save memory,
// only pixelations up to 512 pixels at the equator
// can be defined.
func NewDistMat(pix *Pixelation) (*DistMat, error) {
	rows := pix.Len()
	if pix.Equator()/2 > math.MaxUint8 {
		return nil, fmt.Errorf("pixelation is too large")
	}

	dm := &DistMat{
		pix:  pix,
		rows: rows,
		m:    make([]uint8, rows*rows),
	}

	for px1 := 0; px1 < rows-1; px1++ {
		pt1 := pix.ID(px1).Point()
		for px2 := px1 + 1; px2 < rows; px2++ {
			pt2 := pix.ID(px2).Point()
			d := Distance(pt1, pt2)
			v := uint8(math.Round(d / pix.Step()))

			// The matrix is symmetric
			loc1 := px1*rows + px2
			loc2 := px2*rows + px1
			dm.m[loc1] = v
			dm.m[loc2] = v
		}
	}

	return dm, nil
}

// At returns the value of the ring distance
// between two pixel IDs.
func (dm *DistMat) At(x, y int) int {
	p := x*dm.rows + y
	return int(dm.m[p])
}
