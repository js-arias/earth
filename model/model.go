// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package model implements paleogeographic reconstruction models
// using an isolatitude pixelation
// and discrete time stages.
//
// There are different model types depending
// on how is expected the reconstruction model will be used.
//
// To build a model,
// type PixPlate is used to provide associations between tectonic plates
// (which rotations are defined in a rotation file)
// and the pixels of that plate.
package model

import (
	"fmt"

	"github.com/js-arias/earth"
	"golang.org/x/exp/slices"
)

// A Recons is an editable paleogeographic reconstruction model
// based on an isolatitude pixelation
// and using discrete time stages.
//
// The model is based on tectonic plates
// so a time pixel should be retrieved by its plate.
type Recons struct {
	pix    *earth.Pixelation
	plates map[int]*recPlate
}

// NewRecons creates a new reconstruction model
// with a given pixelation.
func NewRecons(pix *earth.Pixelation) *Recons {
	return &Recons{
		pix:    pix,
		plates: make(map[int]*recPlate),
	}
}

// Add adds a set of pixel locations
// at a time stage,
// in years,
// of a plate.
// Locations is a map in which the key is the pixel ID at the present time,
// and the value is an slice of pixel IDs of the locations
// of the pixel at the time stage
// (because the pixelation is a discrete representation
// of the continuous space,
// some reconstructions will produce multiple destinations for the same pixel).
func (rec *Recons) Add(plate int, locations map[int][]int, age int64) {
	p, ok := rec.plates[plate]
	if !ok {
		p = &recPlate{
			plate: plate,
			pix:   make(map[int]*pixStage),
		}
		rec.plates[plate] = p
	}

	for pixel, stPix := range locations {
		if pixel >= rec.pix.Len() {
			msg := fmt.Errorf("pixel ID %d is invalid", pixel)
			panic(msg)
		}

		px, ok := p.pix[pixel]
		if !ok {
			px = &pixStage{
				id:     pixel,
				stages: make(map[int64][]int),
			}
			p.pix[pixel] = px
		}

		// Set used pixels
		// so every pixel at a time stage will be stored
		// a single time
		rot := px.stages[age]
		used := make(map[int]bool, len(rot)+len(stPix))
		for _, id := range rot {
			used[id] = true
		}

		// add the pixels
		for _, id := range stPix {
			if used[id] {
				continue
			}
			used[id] = true
			rot = append(rot, id)
		}
		slices.Sort(rot)
		px.stages[age] = rot
	}
}

// Pixels returns the pixel IDs of a plate
// at present time.
func (rec *Recons) Pixels(plate int) []int {
	p, ok := rec.plates[plate]
	if !ok {
		return nil
	}

	pxs := make([]int, 0, len(p.pix))
	for _, px := range p.pix {
		pxs = append(pxs, px.id)
	}
	slices.Sort(pxs)
	return pxs
}

// PixStage returns pixel locations at a time stage,
// in years.
// Locations is a map in which the key is the pixel ID at the present time,
// and the value is an slice of pixel IDs of the locations
// of the pixel at the time stage.
func (rec *Recons) PixStage(plate int, age int64) map[int][]int {
	p, ok := rec.plates[plate]
	if !ok {
		return nil
	}

	st := make(map[int][]int, len(p.pix))
	for _, pix := range p.pix {
		sp := pix.stages[age]
		if len(sp) == 0 {
			continue
		}
		st[pix.id] = sp
	}
	return st
}

// Plates returns an slice with the plate IDs
// of the reconstruction model.
func (rec *Recons) Plates() []int {
	ps := make([]int, 0, len(rec.plates))
	for _, p := range rec.plates {
		ps = append(ps, p.plate)
	}
	slices.Sort(ps)
	return ps
}

// Stages returns the time stages,
// in years,
// defined for a reconstruction model.
func (rec *Recons) Stages() []int64 {
	ages := make(map[int64]bool)
	for _, p := range rec.plates {
		for _, pix := range p.pix {
			for a := range pix.stages {
				ages[a] = true
			}
		}
	}

	st := make([]int64, 0, len(ages))
	for a := range ages {
		st = append(st, a)
	}
	slices.Sort(st)

	return st
}

// RecPlate is a collection of time pixels
// associated with a tectonic plate.
type recPlate struct {
	plate int

	// pix is a map of pixels
	// at present time stage
	// to stage pixels.
	pix map[int]*pixStage
}

// A PixStage is a pixel with associated time stages.
type pixStage struct {
	// id is the ID of the pixel at present time
	id int

	// stages store locations at different time stages
	// for a pixel.
	stages map[int64][]int
}
