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
// only pixelations up to 255 pixels at the equator
// can be defined.
func NewDistMat(pix *Pixelation) (*DistMat, error) {
	rows := pix.Len()
	if pix.Equator()/2 > math.MaxUint8 {
		return nil, fmt.Errorf("pixelation is too large")
	}

	dm := &DistMat{
		pix:  pix,
		rows: rows,
		m:    make([]uint8, sizeMatrix(rows)),
	}

	for px1 := 0; px1 < pix.Len(); px1++ {
		pt1 := pix.ID(px1).Point()
		for px2 := 0; px2 <= px1; px2++ {
			pt2 := pix.ID(px2).Point()
			d := Distance(pt1, pt2)
			v := uint8(math.Round(d / ToRad(pix.Step())))

			loc := sizeMatrix(px1) + px2
			dm.m[loc] = v
		}
	}

	return dm, nil
}

// SizeMatrix returns the size of a triangular matrix.
func sizeMatrix(d int) int {
	return (d + 1) * d / 2
}

// At returns the value of the ring distance
// between two pixel IDs.
func (dm *DistMat) At(x, y int) int {
	if y > x {
		x, y = y, x
	}
	loc := sizeMatrix(x) + y
	return int(dm.m[loc])
}
