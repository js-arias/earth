// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package model

import (
	"fmt"

	"github.com/js-arias/earth"
	"golang.org/x/exp/slices"
)

// A TimePix is a pixelated set of values
// (for example,
// an environmental variable)
// at different time stages.
// Note that only positions and values are stored
// so the identity of the pixel in time
// is not preserved.
type TimePix struct {
	pix *earth.Pixelation

	// Pixel values at different time stages
	stages map[int64]*timePix
}

// NewTimePix returns a new time pixelation
// based on an isolatitude pixelation.
func NewTimePix(pix *earth.Pixelation) *TimePix {
	return &TimePix{
		pix:    pix,
		stages: make(map[int64]*timePix),
	}
}

// At returns the value for a pixel at a time
// in a time pixelation.
// If the pixel was never defined,
// it will return the default value
// (i.e. 0).
//
// If the time stage is not defined for the time pixelation
// if will return 0 and false.
// If a pixel value in the closer time stage is wanted,
// use AtCloser.
func (tp *TimePix) At(age int64, pixel int) (int, bool) {
	st, ok := tp.stages[age]
	if !ok {
		return 0, false
	}

	v := st.values[pixel]
	return v, true
}

// AtCloser returns the value for a pixel at the closer time stage
// (i.e. the age of the oldest stage
// younger than the indicated age).
// If the pixel was never defined,
// it will return the default value
// (i.e. 0).
func (tp *TimePix) AtCloser(age int64, pixel int) int {
	age = tp.CloserStageAge(age)
	v, _ := tp.At(age, pixel)
	return v
}

// CloserStageAge returns the closer stage age
// for a time
// (i.e. the age of the oldest stage
// younger than the indicated age).
func (tp *TimePix) CloserStageAge(age int64) int64 {
	st := tp.Stages()
	if i, ok := slices.BinarySearch(st, age); !ok {
		age = st[i-1]
	}
	return age
}

// Pixelation returns the underlying isolatitude pixelation.
func (tp *TimePix) Pixelation() *earth.Pixelation {
	return tp.pix
}

// Set sets a value for a pixel at a time
// in a time pixelation.
func (tp *TimePix) Set(age int64, pixel, value int) {
	if pixel >= tp.pix.Len() {
		msg := fmt.Sprintf("pixel ID %d is invalid", pixel)
		panic(msg)
	}

	st := tp.stages[age]
	if st == nil {
		st = &timePix{
			age:    age,
			values: make(map[int]int),
		}
		tp.stages[age] = st
	}
	st.values[pixel] = value
}

// Stages returns the time stages defined
// for a time pixelation.
func (tp *TimePix) Stages() []int64 {
	st := make([]int64, 0, len(tp.stages))
	for _, a := range tp.stages {
		st = append(st, a.age)
	}
	slices.Sort(st)

	return st
}

type timePix struct {
	// Age of the pixelation
	age int64

	// Value is a map of a pixel ID in the pixelation
	// to an integer value.
	values map[int]int
}
