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
	"sync"
	"time"

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

var pixHead = []string{
	"equator",
	"plate",
	"pixel",
	"begin",
	"end",
}

// ReadPixPlate reads a tectonic plates pixelation
// from a TSV file.
//
// The TSV file must have the following columns:
//
//   - equator, for the number of pixels at the equator
//   - plate, the ID of a tectonic plate
//   - pixel, the ID of a pixel (from a pixelation)
//   - begin, the oldest age of the pixel (in years)
//   - end, the youngest age of the pixel (in years)
//
// Optionally,
// it can include the following fields:
//
//   - name, name of the tectonic feature
//
// Here is an example file:
//
//	# tectonic plates pixelation
//	equator	plate	pixel	name	begin	end
//	360	202	29611	Parana	600000000	0
//	360	802	41257	Antarctica	600000000	0
//
// If no pixelation is given,
// a new pixelation will be created.
func ReadPixPlate(r io.Reader, pix *earth.Pixelation) (*PixPlate, error) {
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
	for _, h := range pixHead {
		if _, ok := fields[h]; !ok {
			return nil, fmt.Errorf("expecting field %q", h)
		}
	}

	var pp *PixPlate
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
			return nil, fmt.Errorf("on row %d: field %q: got %d, want %d", ln, f, eq, pix.Equator())
		}
		if pp == nil {
			pp = NewPixPlate(pix)
		}

		f = "plate"
		plate, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		p := pp.pixPlate(plate)

		f = "pixel"
		id, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		if id >= pix.Len() {
			return nil, fmt.Errorf("on row %d: field %q: invalid pixel value %d", ln, f, id)
		}

		f = "begin"
		begin, err := strconv.ParseInt(row[fields[f]], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		f = "end"
		end, err := strconv.ParseInt(row[fields[f]], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		if end > begin {
			return nil, fmt.Errorf("on row %d: field %q: end value must be less than %d", ln, f, begin)
		}

		name := ""
		f = "name"
		if _, ok := fields[f]; ok {
			name = row[fields[f]]
		}

		p.add(id, name, begin, end)
	}
	if pp == nil {
		return nil, fmt.Errorf("while reading data: %v", io.EOF)
	}
	return pp, nil
}

// TSV encodes a plate pixelation
// into a TSV file.
func (pp *PixPlate) TSV(w io.Writer) error {
	bw := bufio.NewWriter(w)
	fmt.Fprintf(bw, "# tectonic plates pixelation\n")
	fmt.Fprintf(bw, "# data save on: %s\n", time.Now().Format(time.RFC3339))

	tab := csv.NewWriter(bw)
	tab.Comma = '\t'
	tab.UseCRLF = true

	header := []string{
		"equator",
		"plate",
		"pixel",
		"name",
		"begin",
		"end",
	}
	if err := tab.Write(header); err != nil {
		return fmt.Errorf("while writing header: %v", err)
	}

	eq := strconv.Itoa(pp.pix.Equator())

	pp.mu.Lock()
	defer pp.mu.Unlock()

	plates := make([]int, 0, len(pp.plates))
	for _, p := range pp.plates {
		plates = append(plates, p.plate)
	}
	slices.Sort(plates)

	for _, plate := range plates {
		p := pp.plates[plate]
		pxs := make([]int, 0, len(p.pix))
		for _, px := range p.pix {
			pxs = append(pxs, px.ID)
		}
		slices.Sort(pxs)

		pID := strconv.Itoa(plate)

		for _, id := range pxs {
			px := p.pix[id]
			row := []string{
				eq,
				pID,
				strconv.Itoa(id),
				px.Name,
				strconv.FormatInt(px.Begin, 10),
				strconv.FormatInt(px.End, 10),
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
