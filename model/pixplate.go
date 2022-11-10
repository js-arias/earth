// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package model

import (
	"fmt"
	"sync"

	"github.com/js-arias/earth"
	"golang.org/x/exp/slices"
)

// A PixPlate is a collection of pixels
// associated to tectonic plates.
type PixPlate struct {
	pix *earth.Pixelation

	mu     sync.RWMutex
	plates map[int]*pixPlate
}

// NewPixPlate creates a new plate pixelation
// from an isolatitude pixelation.
func NewPixPlate(pix *earth.Pixelation) *PixPlate {
	return &PixPlate{
		pix:    pix,
		plates: make(map[int]*pixPlate),
	}
}

// Add adds a geographic location to a plate
// in a given time frame
// (in years).
func (pp *PixPlate) Add(plate int, name string, lat, lon float64, begin, end int64) {
	p := pp.pixPlate(plate)

	p.mu.Lock()
	p.add(pp.pix.Pixel(lat, lon).ID(), name, begin, end)
	p.mu.Unlock()
}

// AddPixels adds an slice of pixel IDs
// to a plate
// in a given time frame
// (in years).
func (pp *PixPlate) AddPixels(plate int, name string, pixels []int, begin, end int64) {
	p := pp.pixPlate(plate)

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, id := range pixels {
		if id >= pp.pix.Len() {
			msg := fmt.Errorf("pixel ID %d is invalid", id)
			panic(msg)
		}
		p.add(id, name, begin, end)
	}
}

// Pixelation returns the underlying pixelation
// of the pixel collection.
func (pp *PixPlate) Pixelation() *earth.Pixelation {
	return pp.pix
}

// Pixel returns a pixel with of a plate
// with the indicated ID.
func (pp *PixPlate) Pixel(plate, pixel int) PixAge {
	pp.mu.RLock()
	p, ok := pp.plates[plate]
	pp.mu.RUnlock()

	if !ok {
		return PixAge{}
	}

	p.mu.RLock()
	px, ok := p.pix[pixel]
	p.mu.RUnlock()

	if !ok {
		return PixAge{}
	}
	return *px
}

// Pixels return the pixel IDs of a plate.
func (pp *PixPlate) Pixels(plate int) []int {
	pp.mu.RLock()
	p, ok := pp.plates[plate]
	pp.mu.RUnlock()

	if !ok {
		return nil
	}

	p.mu.RLock()
	pxs := make([]int, 0, len(p.pix))
	for _, px := range p.pix {
		pxs = append(pxs, px.ID)
	}
	p.mu.RUnlock()

	slices.Sort(pxs)
	return pxs
}

// Plates return an slice with plate IDs
// of a plate pixelation.
func (pp *PixPlate) Plates() []int {
	pp.mu.RLock()
	defer pp.mu.RUnlock()

	p := make([]int, 0, len(pp.plates))
	for _, pxp := range pp.plates {
		p = append(p, pxp.plate)
	}
	slices.Sort(p)

	return p
}

func (pp *PixPlate) pixPlate(plate int) *pixPlate {
	pp.mu.RLock()
	p, ok := pp.plates[plate]
	pp.mu.RUnlock()

	if ok {
		return p
	}

	p = &pixPlate{
		plate: plate,
		pix:   make(map[int]*PixAge),
	}

	pp.mu.Lock()
	pp.plates[plate] = p
	pp.mu.Unlock()

	return p
}

// A PixAge is a pixel with a defined time range.
type PixAge struct {
	// Name of the feature that contains the pixel
	Name string

	// ID of the pixel
	// (using an isolatitude pixelation)
	ID int

	// Plate ID of the plate that contains the pixel
	Plate int

	// Temporal range of the pixel
	// in years
	Begin int64
	End   int64
}

// A pixPlate is a collection of pixels
// associated with a particular tectonic plate.
type pixPlate struct {
	plate int // ID of the plate

	mu  sync.RWMutex
	pix map[int]*PixAge
}

func (pp *pixPlate) add(id int, name string, begin, end int64) {
	px, ok := pp.pix[id]
	if !ok {
		px = &PixAge{
			Name:  name,
			ID:    id,
			Plate: pp.plate,
			Begin: begin,
			End:   end,
		}
		pp.pix[id] = px
		return
	}

	if px.Name == "" {
		px.Name = name
	}
	// set younger date for the end time
	if px.End > end {
		px.End = end
	}

	// set older date for the begin time
	if px.Begin < begin {
		px.Begin = begin
		if name != "" {
			px.Name = name
		}
	}
}
