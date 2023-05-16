// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package model

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

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
// based on an equal area pixelation.
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

// AtClosest returns the value for a pixel at the closest time stage
// (i.e. the age of the oldest stage
// younger than the indicated age).
// If the pixel was never defined,
// it will return the default value
// (i.e. 0).
func (tp *TimePix) AtClosest(age int64, pixel int) int {
	age = tp.ClosestStageAge(age)
	v, _ := tp.At(age, pixel)
	return v
}

// Bounds return the age bounds for the stage of the given age
// in million years.
func (tp *TimePix) Bounds(age int64) (old, young int64) {
	st := tp.Stages()
	i, ok := slices.BinarySearch(st, age)
	if !ok {
		i = i - 1
	}
	if i+1 >= len(st) {
		return earth.Age, st[i]
	}
	return st[i+1], st[i]
}

// ClosestStageAge returns the closest stage age
// for a time
// (i.e. the age of the oldest stage
// younger than the indicated age).
func (tp *TimePix) ClosestStageAge(age int64) int64 {
	st := tp.Stages()
	if i, ok := slices.BinarySearch(st, age); !ok {
		age = st[i-1]
	}
	return age
}

// Del removes a pixel value at a time
// in a time pixelation.
func (tp *TimePix) Del(age int64, pixel int) {
	if pixel >= tp.pix.Len() {
		return
	}

	st, ok := tp.stages[age]
	if !ok {
		return
	}
	delete(st.values, pixel)
}

// Pixelation returns the underlying equal area pixelation.
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

// Stage returns the values for all pixels
// at a given age
// (in years).
func (tp *TimePix) Stage(age int64) map[int]int {
	st, ok := tp.stages[age]
	if !ok {
		return nil
	}

	return st.values
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

var tpHeader = []string{
	"equator",
	"age",
	"stage-pixel",
	"value",
}

// ReadTimePix reads values of a time pixelation
// from a TSV file.
//
// The TSV must contain the following columns:
//
//   - equator, for the number of pixels at the equator
//   - age, the age of the time stage (in years)
//   - stage-pixel, the pixel ID at the time stage
//   - value, an integer value
//
// Here is an example file:
//
//	equator	age	stage-pixel	value
//	360	100000000	19051	1
//	360	100000000	19055	2
//	360	100000000	19409	1
//	360	140000000	20051	1
//	360	140000000	20055	2
//	360	140000000	20056	3
//
// If no pixelation is given,
// a new pixelation will be created.
func ReadTimePix(r io.Reader, pix *earth.Pixelation) (*TimePix, error) {
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
	for _, h := range tpHeader {
		if _, ok := fields[h]; !ok {
			return nil, fmt.Errorf("expecting field %q", h)
		}
	}

	var tp *TimePix
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
		if tp == nil {
			tp = NewTimePix(pix)
		}

		f = "age"
		age, err := strconv.ParseInt(row[fields[f]], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		st := tp.stages[age]
		if st == nil {
			st = &timePix{
				age:    age,
				values: make(map[int]int),
			}
			tp.stages[age] = st
		}

		f = "stage-pixel"
		px, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		if px >= pix.Len() {
			return nil, fmt.Errorf("on row %d: field %q: invalid pixel value %d", ln, f, px)
		}

		f = "value"
		v, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		st.values[px] = v
	}

	if tp == nil {
		return nil, fmt.Errorf("while reading data: %v", io.EOF)
	}
	return tp, nil
}

// TSV encodes a time pixelation
// as a TSV file.
func (tp *TimePix) TSV(w io.Writer) error {
	bw := bufio.NewWriter(w)
	fmt.Fprintf(bw, "# time pixelation values\n")
	fmt.Fprintf(bw, "# data save on: %s\n", time.Now().Format(time.RFC3339))
	tab := csv.NewWriter(bw)
	tab.Comma = '\t'
	tab.UseCRLF = true

	if err := tab.Write(tpHeader); err != nil {
		return fmt.Errorf("while writing header: %v", err)
	}

	eq := strconv.Itoa(tp.pix.Equator())

	ages := tp.Stages()
	for _, a := range ages {
		age := strconv.FormatInt(a, 10)
		st := tp.stages[a]

		pxs := make([]int, 0, len(st.values))
		for id := range st.values {
			pxs = append(pxs, id)
		}
		slices.Sort(pxs)

		for _, id := range pxs {
			row := []string{
				eq,
				age,
				strconv.Itoa(id),
				strconv.Itoa(st.values[id]),
			}
			if err := tab.Write(row); err != nil {
				return fmt.Errorf("while writing data: %v", err)
			}
		}
	}

	tab.Flush()
	if err := tab.Error(); err != nil {
		return fmt.Errorf("while writing data: %v", err)
	}
	if err := bw.Flush(); err != nil {
		return fmt.Errorf("while writing data: %v", err)
	}
	return nil
}
