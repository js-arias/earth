// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package model

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/js-arias/earth"
	"golang.org/x/exp/slices"
)

// Total is a collection of total rotations
// for pixels at present time
// moved to a given time stage.
type Total struct {
	inverse bool

	pix *earth.Pixelation

	// Reconstructed stages
	stages map[int64]*Rotation
}

// NewTotal returns a collection of total rotations
// build from a tectonic reconstruction model.
func NewTotal(rec *Recons) *Total {
	st := rec.Stages()

	t := &Total{
		pix:    rec.Pixelation(),
		stages: make(map[int64]*Rotation),
	}

	plates := rec.Plates()
	for _, a := range st {
		rot := &Rotation{
			From: 0,
			To:   a,
			Rot:  make(map[int][]int),
		}
		for _, p := range plates {
			sp := rec.PixStage(p, a)
			for from, to := range sp {
				rot.Rot[from] = append(rot.Rot[from], to...)
			}
		}
		t.stages[a] = rot
	}

	// Remove duplicated pixels
	// if any
	for _, rot := range t.stages {
		rot.removeDuplicates()
	}

	return t
}

// ReadTotal reads a collection of total rotations
// from a TSV file that contains
// a tectonic reconstruction model.
// A total rotation is a rotation of a pixel in present time
// to a given time stage.
//
// If no pixelation is given
// a new pixelation will be created.
//
// If inverse is true,
// an inverse rotation will be returned.
// In an inverse rotation,
// the reference pixel are at the given time stage
// an rotate it to their present locations.
func ReadTotal(r io.Reader, pix *earth.Pixelation, inverse bool) (*Total, error) {
	tab := csv.NewReader(r)
	tab.Comma = '\t'
	tab.Comment = '#'

	head, err := tab.Read()
	if err != nil {
		return nil, fmt.Errorf("while reading header: %v", err)
	}
	fields := make(map[string]int, len(head))
	for i, h := range head {
		h = strings.ToLower(h)
		fields[h] = i
	}
	for _, h := range recHeader {
		if _, ok := fields[h]; !ok {
			return nil, fmt.Errorf("expecting field %q", h)
		}
	}

	var tot *Total
	for {
		row, err := tab.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		ln, _ := tab.FieldPos(0)
		if err != nil {
			return nil, fmt.Errorf("on row %d: %v", ln, err)
		}

		f := "equator"
		eq, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		if pix == nil {
			pix = earth.NewPixelation(eq)
		}
		if pix.Equator() != eq {
			return nil, fmt.Errorf("on row %d: field %q: got %d, want %d value", ln, f, eq, pix.Equator())
		}
		if tot == nil {
			tot = &Total{
				pix:     pix,
				stages:  make(map[int64]*Rotation),
				inverse: inverse,
			}
		}

		f = "age"
		age, err := strconv.ParseInt(row[fields[f]], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}

		rot, ok := tot.stages[age]
		if !ok {
			rot = &Rotation{
				From: 0,
				To:   age,
				Rot:  make(map[int][]int),
			}
			if inverse {
				rot.From = age
				rot.To = 0
			}
			tot.stages[age] = rot
		}

		f = "pixel"
		id, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		if id >= pix.Len() {
			return nil, fmt.Errorf("on row %d: field %q: invalid pixel value %d", ln, f, id)
		}
		f = "stage-pixel"
		sID, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		if sID >= pix.Len() {
			return nil, fmt.Errorf("on row %d: field %q: invalid pixel value %d", ln, f, sID)
		}
		if inverse {
			rot.Rot[sID] = append(rot.Rot[sID], id)
		} else {
			rot.Rot[id] = append(rot.Rot[id], sID)
		}
	}
	if tot == nil {
		return nil, fmt.Errorf("while reading data: %v", io.EOF)
	}

	// Remove duplicated pixels
	// if any
	for _, rot := range tot.stages {
		rot.removeDuplicates()
	}

	return tot, nil
}

// ClosesStageAge returns the closer stage age
// for a given time age
// (i.e. the age of the oldest time stage
// that is youngest than the given age).
// This stage age is the one used by Rotation method.
func (t *Total) ClosesStageAge(age int64) int64 {
	st := t.Stages()
	if i, ok := slices.BinarySearch(st, age); !ok {
		age = st[i-1]
	}
	return age
}

// Inverse returns an inverse total rotation,
// a collection of pixels in past time
// moved to current time.
func (t *Total) Inverse() *Total {
	st := t.Stages()

	inv := &Total{
		inverse: true,
		pix:     t.pix,
		stages:  make(map[int64]*Rotation),
	}

	for _, a := range st {
		rot := &Rotation{
			From: a,
			To:   0,
			Rot:  make(map[int][]int),
		}
		tot := t.Rotation(a)
		for id, v := range tot {
			for _, px := range v {
				rot.Rot[px] = append(rot.Rot[px], id)
			}
		}
		inv.stages[a] = rot
	}

	// Remove duplicated pixels
	// if any
	for _, rot := range inv.stages {
		rot.removeDuplicates()
	}

	return inv
}

// IsInverse returns true in the total rotation
// is inverse
// i.e. is from past pixels to present pixels.
func (t *Total) IsInverse() bool {
	return t.inverse
}

// Pixelation returns the underlying pixelation
// of a total rotation model.
func (t *Total) Pixelation() *earth.Pixelation {
	return t.pix
}

// Rotation returns a pixel location at a given time stage.
// Locations is a map in which the key is the pixel ID at present time,
// and the value is an slice of pixel IDs of the locations
// of the key pixel at the time stage.
//
// If the age given is not a defined time stage,
// the returned locations will be from the oldest time stage
// that is youngest that the given age.
// For example,
// the defined stages are [0, 10_000_000, 100_000_000],
// if asked for the stage 19_843_211
// it will return the pixel locations at 10_000_000.
func (t *Total) Rotation(age int64) map[int][]int {
	age = t.ClosesStageAge(age)

	rot := t.stages[age]
	return rot.Rot
}

// Stages return the time stages defined
// for the total rotation model.
func (t *Total) Stages() []int64 {
	st := make([]int64, 0, len(t.stages))
	for _, rot := range t.stages {
		if t.inverse {
			st = append(st, rot.From)
			continue
		}
		st = append(st, rot.To)
	}
	slices.Sort(st)

	return st
}

// A Rotation is a rotation of a pixel
// to another time stage.
type Rotation struct {
	// Ages (in years) of the rotation
	From int64
	To   int64

	// pixels at 'from' time rotate to 'to' time
	Rot map[int][]int
}

func (r *Rotation) removeDuplicates() {
	for px, dest := range r.Rot {
		used := make(map[int]bool, len(dest))
		for _, id := range dest {
			used[id] = true
		}

		pix := make([]int, 0, len(used))
		for id := range used {
			pix = append(pix, id)
		}
		slices.Sort(pix)
		r.Rot[px] = pix
	}
}
