// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package earth

import "math"

// A Network is an slice of pixel IDs
// to the ID of their closest pixel neighbors
// (including the original pixel).
type Network [][]int

// NewNetwork returns a new network
// from a given pixelation.
func NewNetwork(pix *Pixelation) Network {
	r := ToRad(pix.dStep) * math.Sqrt2

	net := make(Network, pix.Len())
	for px1 := range net {
		px := pix.ID(px1)
		pt1 := px.Point()
		start := px.Ring() - 1
		if start < 0 {
			start = 0
		}
		end := px.Ring() + 1
		var n []int

		for px2 := pix.FirstPix(start).ID(); px2 < pix.Len(); px2++ {
			op := pix.ID(px2)
			if op.Ring() > end {
				break
			}
			if px1 == px2 {
				n = append(n, px2)
				continue
			}
			pt2 := op.Point()
			if Distance(pt1, pt2) < r {
				n = append(n, px2)
			}
		}
		net[px1] = n
	}

	return net
}
