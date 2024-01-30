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
// but a distance suing an integer scale.
type DistMat struct {
	pix   *Pixelation
	rows  int
	scale uint16
	m     []uint16
}

// NewDistMat creates a new distance matrix
// from the indicated pixelation.
// To save memory,
// only pixelations up to 255 pixels at the equator
// can be defined.
func NewDistMat(pix *Pixelation, scale uint16) (*DistMat, error) {
	rows := pix.Len()
	if pix.Equator()/2 > math.MaxUint8 {
		return nil, fmt.Errorf("pixelation is too large")
	}

	dm := &DistMat{
		pix:   pix,
		rows:  rows,
		scale: scale,
		m:     make([]uint16, sizeMatrix(rows)),
	}

	for px1 := 0; px1 < pix.Len(); px1++ {
		pt1 := pix.ID(px1).Point()
		for px2 := 0; px2 <= px1; px2++ {
			pt2 := pix.ID(px2).Point()
			d := Distance(pt1, pt2)
			v := uint16(math.Round(d * float64(scale) / math.Pi))

			loc := sizeMatrix(px1) + px2
			dm.m[loc] = v
		}
	}

	return dm, nil
}

// NewDistMatRingScale returns a new distance matrix
// from the indicated pixelations,
// scaled with the number of rings in the pixelation.
// Then the distance is equal to the ring of each pixel
// if a reference pixel is rotated to the north pole.
func NewDistMatRingScale(pix *Pixelation) (*DistMat, error) {
	return NewDistMat(pix, uint16(pix.Rings()-1))
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

// Scale returns the scale factor used in the distance matrix.
func (dm *DistMat) Scale() int {
	return int(dm.scale)
}

// SizeMatrix returns the size of a triangular matrix.
func sizeMatrix(d int) int {
	return (d + 1) * d / 2
}
