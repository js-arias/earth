// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package weight associates a pixelation raster value
// with a normalized weight (between 0 and 1) for a pixel.
package weight

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Weight stores the normalized weight
// of a pixel.
type weight struct {
	w  float64
	ln float64 // logWeight
}

// Pixel is a set of normalized weights
// (a value between 0 and 1)
// for the values given to a pixel.
//
// Each pixel is assumed to be independent
// of all other pixels.
type Pixel map[int]weight

// New creates a new Pixel object to store
// normalized weight from pixel types.
//
// By default,
// the ID 0 is defined with weight 0.
func New() Pixel {
	pw := map[int]weight{
		0: {w: 0, ln: math.Inf(-1)},
	}
	return pw
}

// LogWeight returns the log of the weight
// for a given raster value.
func (px Pixel) LogWeight(v int) float64 {
	p, ok := px[v]
	if !ok {
		return math.Inf(-1)
	}
	return p.ln
}

// Weight returns the normalized weight
// for a given raster value.
func (px Pixel) Weight(v int) float64 {
	p, ok := px[v]
	if !ok {
		return 0
	}
	return p.w
}

// Set set a pixel normalized weight
// for a given raster value.
func (px Pixel) Set(v int, w float64) error {
	if w < 0 || w > 1 {
		return fmt.Errorf("invalid weight value %.6f", w)
	}
	px[v] = weight{
		w:  w,
		ln: math.Log(w),
	}
	return nil
}

// Values return the raster values
// that have a defined weights.
func (px Pixel) Values() []int {
	vs := make([]int, 0, len(px))
	for v := range px {
		vs = append(vs, v)
	}
	slices.Sort(vs)

	return vs
}

// TSV encodes pixel weights as a TSV file.
func (px Pixel) TSV(w io.Writer) error {
	for k, p := range px {
		if p.w < 0 || p.w > 1 {
			return fmt.Errorf("invalid pixel weight %.6f for pixel %d", p.w, k)
		}
	}

	bw := bufio.NewWriter(w)
	fmt.Fprintf(bw, "# normalized pixel weights\n")
	fmt.Fprintf(bw, "# data save on: %s\n", time.Now().Format(time.RFC3339))
	tab := csv.NewWriter(bw)
	tab.Comma = '\t'
	tab.UseCRLF = true
	if err := tab.Write([]string{"key", "weight"}); err != nil {
		return fmt.Errorf("while writing header: %v", err)
	}

	vs := px.Values()
	for _, v := range vs {
		row := []string{
			strconv.Itoa(v),
			strconv.FormatFloat(px[v].w, 'f', 6, 64),
		}
		if err := tab.Write(row); err != nil {
			return fmt.Errorf("while writing data: %v", err)
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

// ReadTSV reads a TSV file used to define the normalized weight values
// for a given set of pixels values in a pixelation.
//
// The pixel weight file is a tab-delimited file
// with the following columns:
//
//	-key	the value used as identifier
//	-weight	the normalized weight for a pixel with that value
//
// Any other columns,
// will be ignored.
// Here is an example of a pixel weight file:
//
//	key	weight	comment
//	0	0.000000	deep ocean
//	1	0.010000	oceanic plateaus
//	2	0.050000	continental shelf
//	3	0.950000	lowlands
//	4	1.000000	highlands
//	5	0.001000	ice sheets
func ReadTSV(r io.Reader) (Pixel, error) {
	tsv := csv.NewReader(r)
	tsv.Comma = '\t'
	tsv.Comment = '#'

	head, err := tsv.Read()
	if err != nil {
		return nil, fmt.Errorf("while reading header: %v", err)
	}
	fields := make(map[string]int, len(head))
	for i, h := range head {
		h = strings.ToLower(h)
		fields[h] = i
	}
	for _, h := range []string{"key", "weight"} {
		if _, ok := fields[h]; !ok {
			return nil, fmt.Errorf("expecting field %q", h)
		}
	}

	p := New()
	for {
		row, err := tsv.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		ln, _ := tsv.FieldPos(0)
		if err != nil {
			return nil, fmt.Errorf("on row %d: %v", ln, err)
		}

		f := "key"
		k, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}

		f = "weight"
		w, err := strconv.ParseFloat(row[fields[f]], 64)
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		if w < 0 || w > 1 {
			return nil, fmt.Errorf("on row %d: field %q: invalid weight value %.6f", ln, f, w)
		}

		p[k] = weight{
			w:  w,
			ln: math.Log(w),
		}
	}

	return p, nil
}
