// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package model

import (
	"io"

	"github.com/js-arias/earth"
	"golang.org/x/exp/slices"
)

// StageRot is a collection of stage rotations.
//
// An stage rotation is a rotation from a pixel
// at a time stage to another time stage.
// In this case the time stages
// are the neighbor time stages
// defined by a tectonic reconstruction model.
type StageRot struct {
	pix *earth.Pixelation

	// This rotations contains the pixel identities
	// from a stage to a neighbor stage.
	//
	// -YoungToOld contains the pixels at a given stage,
	//	mapped to an slice of the destination pixels
	//	in a previous (oldest) stage.
	// - OldToYoung contains the pixels at a given stage,
	//	mapped to an slice of the destination pixels
	//	in a posterior (youngest) stage.
	youngToOld map[int64]*Rotation
	oldToYoung map[int64]*Rotation
}

// NewStageRot returns a collection of stage rotations
// from a reconstruction model.
func NewStageRot(rec *Recons) *StageRot {
	st := rec.Stages()

	s := &StageRot{
		pix:        rec.pix,
		youngToOld: make(map[int64]*Rotation),
		oldToYoung: make(map[int64]*Rotation),
	}

	// make the stage rotations
	for _, p := range rec.Plates() {
		for i, a := range st {
			if i == 0 {
				continue
			}
			y := st[i-1]

			old := rec.PixStage(p, a)
			if old == nil {
				continue
			}

			young := rec.PixStage(p, y)
			if young == nil {
				continue
			}

			// Rotation from an older stage
			// to a younger stage
			o2y, ok := s.oldToYoung[a]
			if !ok {
				o2y = &Rotation{
					From: a,
					To:   y,
					Rot:  make(map[int][]int),
				}
				s.oldToYoung[a] = o2y
			}
			for pp, v := range old {
				for _, px := range v {
					o2y.Rot[px] = append(o2y.Rot[px], young[pp]...)
				}
			}

			// Rotation from a younger stage
			// to an older stage
			y2o, ok := s.youngToOld[y]
			if !ok {
				y2o = &Rotation{
					From: y,
					To:   a,
					Rot:  make(map[int][]int),
				}
				s.youngToOld[y] = y2o
			}
			for pp, v := range young {
				for _, px := range v {
					y2o.Rot[px] = append(y2o.Rot[px], old[pp]...)
				}
			}
		}
	}

	// remove duplicates
	for _, o2y := range s.oldToYoung {
		o2y.removeDuplicates()
	}
	for _, y2o := range s.youngToOld {
		y2o.removeDuplicates()
	}

	return s
}

// ReadStageRot reads a collection of stage rotations
// from a TSV file that contains
// a plate motion model.
//
// The TSV file is a paleogeographic reconstruction model
// and must contains the following columns:
//
// The TSV file is a paleogeographic reconstruction model file
// and must contain the following columns:
//
//   - equator, for the number of pixels at the equator
//   - plate, the ID of a tectonic plate
//   - pixel, the ID of a pixel (in the isolatitude pixelation)
//   - age, the age of the time stage (in years)
//   - stage-pixel, the pixel ID at the time stage
//
// Here is an example file:
//
//	# paleogeographic reconstruction model
//	equator	plate	pixel	age	stage-pixel
//	360	59999	17051	100000000	19051
//	360	59999	17051	140000000	20051
//	360	59999	17055	100000000	19055
//	360	59999	17055	140000000	20055
//	360	59999	17055	140000000	20056
//
// If no pixelation is given,
// a new pixelation will be created.
func ReadStageRot(r io.Reader, pix *earth.Pixelation) (*StageRot, error) {
	rec, err := ReadReconsTSV(r, pix)
	if err != nil {
		return nil, err
	}

	return NewStageRot(rec), nil
}

// ClosestStageAge returns the closest stage age
// for a given time
// i.e. the age of the first time stage younger than the given age.
func (s *StageRot) ClosestStageAge(age int64) int64 {
	st := s.Stages()
	if i, ok := slices.BinarySearch(st, age); !ok {
		age = st[i-1]
	}
	return age
}

// OldToYoung returns an stage rotation from an older stage
// to it most immediate younger stage.
// If there is no younger stage,
// it will return a nil map.
func (s *StageRot) OldToYoung(oldStage int64) *Rotation {
	o2y, ok := s.oldToYoung[oldStage]
	if !ok {
		return nil
	}
	return o2y
}

// Stages return the time stages defined
// for the stage rotation model.
func (s *StageRot) Stages() []int64 {
	ages := make(map[int64]bool)
	for a := range s.oldToYoung {
		ages[a] = true
	}
	for a := range s.youngToOld {
		ages[a] = true
	}

	st := make([]int64, 0, len(ages))
	for a := range ages {
		st = append(st, a)
	}
	slices.Sort(st)
	return st
}

// YoungToOld returns an stage rotation from a younger stage
// to it most immediate older stage.
// If there is no older stage,
// it will return a nil map.
func (s *StageRot) YoungToOld(youngStage int64) *Rotation {
	y2o, ok := s.youngToOld[youngStage]
	if !ok {
		return nil
	}
	return y2o
}
