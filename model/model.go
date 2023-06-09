// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package model implements paleogeographic reconstruction models
// using a pixelation based on an equal area pixelation
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

// A Recons is an editable plate motion model
// based on an equal area pixelation
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

// Pixelation returns the underlying equal area pixelation
// of the model.
func (rec *Recons) Pixelation() *earth.Pixelation {
	return rec.pix
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

func (ps *pixStage) removeDuplicates() {
	for a, rot := range ps.stages {
		used := make(map[int]bool, len(rot))
		for _, id := range rot {
			used[id] = true
		}

		pix := make([]int, 0, len(used))
		for id := range used {
			pix = append(pix, id)
		}
		slices.Sort(pix)
		ps.stages[a] = pix
	}
}

var recHeader = []string{
	"equator",
	"plate",
	"pixel",
	"age",
	"stage-pixel",
}

// ReadReconsTSV reads a plate motion model
// from a TSV file.
//
// The TSV file must contains the following columns:
//
//   - equator, for the number of pixels at the equator
//   - plate, the ID of a tectonic plate
//   - pixel, the ID of a pixel (in an equal area pixelation)
//   - age, the age of the time stage (in years)
//   - stage-pixel, the pixel ID at the time stage
//
// Here is an example file:
//
//	equator	plate	pixel	age	stage-pixel
//	360	59999	17051	100000000	19051
//	360	59999	17051	140000000	20051
//	360	59999	17055	100000000	19055
//	360	59999	17055	140000000	20055
//	360	59999	17055	140000000	20056
//
// If no pixelation is given,
// a new pixelation will be created.
func ReadReconsTSV(r io.Reader, pix *earth.Pixelation) (*Recons, error) {
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

	var rec *Recons
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
		if rec == nil {
			rec = NewRecons(pix)
		}

		f = "plate"
		plate, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		p, ok := rec.plates[plate]
		if !ok {
			p = &recPlate{
				plate: plate,
				pix:   make(map[int]*pixStage),
			}
			rec.plates[plate] = p
		}

		f = "pixel"
		id, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		if id >= pix.Len() {
			return nil, fmt.Errorf("on row %d: field %q: invalid pixel value %d", ln, f, id)
		}
		px, ok := p.pix[id]
		if !ok {
			px = &pixStage{
				id:     id,
				stages: make(map[int64][]int),
			}
			p.pix[id] = px
		}

		f = "age"
		age, err := strconv.ParseInt(row[fields[f]], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}

		f = "stage-pixel"
		sID, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		if sID >= pix.Len() {
			return nil, fmt.Errorf("on row %d: field %q: invalid pixel value %d", ln, f, sID)
		}
		px.stages[age] = append(px.stages[age], sID)
	}

	if rec == nil {
		return nil, fmt.Errorf("while reading data: %v", io.EOF)
	}

	// Remove duplicated pixels,
	// if any
	for _, plate := range rec.plates {
		for _, px := range plate.pix {
			px.removeDuplicates()
		}
	}

	return rec, nil
}

// TSV encodes a plate motion model
// as a TSV file.
func (rec *Recons) TSV(w io.Writer) error {
	bw := bufio.NewWriter(w)
	fmt.Fprintf(bw, "# plate motion model\n")
	fmt.Fprintf(bw, "# data save on: %s\n", time.Now().Format(time.RFC3339))
	tab := csv.NewWriter(bw)
	tab.Comma = '\t'
	tab.UseCRLF = true

	if err := tab.Write(recHeader); err != nil {
		return fmt.Errorf("while writing header: %v", err)
	}

	eq := strconv.Itoa(rec.pix.Equator())

	plates := make([]int, 0, len(rec.plates))
	for _, p := range rec.plates {
		plates = append(plates, p.plate)
	}
	slices.Sort(plates)

	for _, p := range plates {
		plate := rec.plates[p]
		pxs := make([]int, 0, len(plate.pix))
		for _, px := range plate.pix {
			pxs = append(pxs, px.id)
		}
		slices.Sort(pxs)

		pID := strconv.Itoa(plate.plate)

		for _, id := range pxs {
			ps := plate.pix[id]
			st := make([]int64, 0, len(ps.stages))
			for a := range ps.stages {
				st = append(st, a)
			}
			slices.Sort(st)

			pixID := strconv.Itoa(ps.id)

			for _, a := range st {
				age := strconv.FormatInt(a, 10)
				for _, sp := range ps.stages[a] {
					row := []string{
						eq,
						pID,
						pixID,
						age,
						strconv.Itoa(sp),
					}
					if err := tab.Write(row); err != nil {
						return fmt.Errorf("while writing data: %v", err)
					}
				}
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
