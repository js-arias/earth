// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package pixprob associates a pixelation raster value
// with a probability for a pixel.
package pixprob

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

// Pixel is the prior probability of a pixel
// given a raster value.
//
// Each pixel is assumed to be independent
// of all other pixels.
type Pixel map[int]float64

// New creates a new Pixel object to store
// prior probabilities from pixel types.
//
// By default,
// the ID 0 is defined with probability 0.
func New() Pixel {
	return Pixel{0: 0}
}

// Prior returns the prior probability
// of a pixel for a given raster value.
func (px Pixel) Prior(v int) float64 {
	return px[v]
}

// Set set a pixel probability
// for a given raster value.
func (px Pixel) Set(v int, prob float64) error {
	if prob < 0 || prob > 1 {
		return fmt.Errorf("invalid prior value %.6f", prob)
	}
	px[v] = prob
	return nil
}

// Values return the raster values
// that have a defined prior.
func (px Pixel) Values() []int {
	vs := make([]int, 0, len(px))
	for v := range px {
		vs = append(vs, v)
	}
	slices.Sort(vs)

	return vs
}

// TSV encodes a pixel prior as a TSV file.
func (px Pixel) TSV(w io.Writer) error {
	for k, prob := range px {
		if prob < 0 || prob > 1 {
			return fmt.Errorf("invalid pixel probability %.6f for pixel %d", prob, k)
		}
	}

	bw := bufio.NewWriter(w)
	fmt.Fprintf(bw, "# pixel priors\n")
	fmt.Fprintf(bw, "# data save on: %s\n", time.Now().Format(time.RFC3339))
	tab := csv.NewWriter(bw)
	tab.Comma = '\t'
	tab.UseCRLF = true
	if err := tab.Write([]string{"key", "prior"}); err != nil {
		return fmt.Errorf("while writing header: %v", err)
	}

	vs := px.Values()
	for _, v := range vs {
		row := []string{
			strconv.Itoa(v),
			strconv.FormatFloat(px[v], 'f', 6, 64),
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

// ReadTSV reads a TSV file used to define the prior probability values
// for a given set of pixels values in a pixelation.
//
// The pixel prior file is a tab-delimited file
// with the following columns:
//
//	-key	the value used as identifier
//	-prior	the prior probability for a pixel with that value
//
// Any other columns,
// will be ignored.
// Here is an example of a pixel prior file:
//
//	key	prior	comment
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
	for _, h := range []string{"key", "prior"} {
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

		f = "prior"
		pp, err := strconv.ParseFloat(row[fields[f]], 64)
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		if pp < 0 || pp > 1 {
			return nil, fmt.Errorf("on row %d: field %q: invalid prior value %.6f", ln, f, pp)
		}

		p[k] = pp
	}

	return p, nil
}
