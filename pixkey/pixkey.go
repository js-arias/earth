// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package pixkey implements a key to relate raster values
// in a landscape pixelation
// to pixel labels and a simple color key.
package pixkey

import (
	"encoding/csv"
	"errors"
	"fmt"
	"image/color"
	"io"
	"slices"
	"strconv"
	"strings"
)

// PixKey stores the color values
// for a pixel value.
type PixKey struct {
	values map[string]int
	labels map[int]string

	color map[int]color.Color
	gray  map[int]uint8
}

// New creates a new empty key.
func New() *PixKey {
	return &PixKey{
		values: make(map[string]int),
		labels: make(map[int]string),
		color:  make(map[int]color.Color),
		gray:   make(map[int]uint8),
	}
}

// Color returns the color associated with a given value.
// If no color is defined for the value,
// it will return transparent black.
func (pk *PixKey) Color(v int) (color.Color, bool) {
	c, ok := pk.color[v]
	if !ok {
		return color.RGBA{0, 0, 0, 0}, false
	}
	return c, true
}

// HasGrayScale returns true if a gray scale is defined
// for the keys.
func (pk *PixKey) HasGrayScale() bool {
	return len(pk.gray) > 0
}

// Key returns the key value for a given label.
func (pk *PixKey) Key(label string) int {
	label = strings.Join(strings.Fields(strings.ToLower(label)), " ")
	if label == "" {
		return 0
	}
	return pk.values[label]
}

// Keys return the defined key values.
func (pk *PixKey) Keys() []int {
	keys := make([]int, 0, len(pk.color))
	for k := range pk.color {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// Label returns the label for a given key value.
func (pk *PixKey) Label(v int) string {
	return pk.labels[v]
}

// Gray returns the gray color associated with a given value.
// If no color is defined for the value,
// it will return transparent black.
func (pk *PixKey) Gray(v int) (color.Color, bool) {
	g, ok := pk.gray[v]
	if !ok {
		return color.RGBA{0, 0, 0, 0}, false
	}
	return color.RGBA{g, g, g, 255}, true
}

// SetColor sets a color to be associated with a given value.
func (pk *PixKey) SetColor(c color.Color, v int) {
	pk.color[v] = c
}

// SetGray sets the gray color to be associated with a given value.
func (pk *PixKey) SetGray(c color.Color, v int) {
	r, _, _, _ := c.RGBA()
	pk.gray[v] = uint8(r >> 8)
}

// SetLabel sets the label of a given ket value.
func (pk *PixKey) SetLabel(v int, label string) error {
	label = strings.Join(strings.Fields(strings.ToLower(label)), " ")
	if label == "" {
		return nil
	}

	if _, ok := pk.color[v]; !ok {
		return nil
	}
	l := pk.labels[v]
	if l == label {
		return nil
	}
	if _, ok := pk.values[label]; ok {
		return fmt.Errorf("label %q already in use", label)
	}

	delete(pk.values, l)
	pk.labels[v] = label
	pk.values[label] = v
	return nil
}

// Read reads a key file used to define the colors
// for pixel values in a time pixelation.
//
// A key file is a tab-delimited file
// with the following required columns:
//
//	-key	the value used as identifier
//	-color	an RGB value separated by commas,
//		for example "125,132,148".
//
// Optionally it can contain the following columns:
//
//	-label: for the pixel label
//	-gray:  for a gray scale value
//
// Any other columns, will be ignored.
// Here is an example of a key file:
//
//	key	label	color	gray
//	0	deep ocean	0, 26, 51	0	deep ocean
//	1	oceanic plateaus	0, 84, 119	10
//	2	continental shelf	68, 167, 196	20
//	3	lowlands	251, 236, 93	90
//	4	highlands	255, 165, 0	100
//	5	ice sheets	229, 229, 224	50
func Read(r io.Reader) (*PixKey, error) {
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
	for _, h := range []string{"key", "color"} {
		if _, ok := fields[h]; !ok {
			return nil, fmt.Errorf("expecting field %q", h)
		}
	}

	pk := &PixKey{
		values: make(map[string]int),
		labels: make(map[int]string),
		color:  make(map[int]color.Color),
		gray:   make(map[int]uint8),
	}

	for {
		row, err := tab.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		ln, _ := tab.FieldPos(0)
		if err != nil {
			return nil, fmt.Errorf("on row %d: %v", ln, err)
		}

		f := "key"
		k, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		if _, ok := pk.color[k]; ok {
			return nil, fmt.Errorf("on row %d: field %q: key '%d' already used", ln, f, k)
		}

		f = "color"
		val := strings.Split(row[fields[f]], ",")
		if len(val) != 3 {
			return nil, fmt.Errorf("on row %d: field %q: found %d values, want 3", ln, f, len(val))
		}

		red, err := strconv.Atoi(strings.TrimSpace(val[0]))
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q [red value]: %v", ln, f, err)
		}
		if red > 255 {
			return nil, fmt.Errorf("on row %d: field %q [red value]: invalid value %d", ln, f, red)
		}
		green, err := strconv.Atoi(strings.TrimSpace(val[1]))
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q [green value]: %v", ln, f, err)
		}
		if green > 255 {
			return nil, fmt.Errorf("on row %d: field %q [green value]: invalid value %d", ln, f, green)
		}
		blue, err := strconv.Atoi(strings.TrimSpace(val[2]))
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q [blue value]: %v", ln, f, err)
		}
		if blue > 255 {
			return nil, fmt.Errorf("on row %d: field %q [blue value]: invalid value %d", ln, f, blue)
		}

		c := color.RGBA{uint8(red), uint8(green), uint8(blue), 255}
		pk.color[k] = c

		label := strconv.Itoa(k)
		f = "label"
		if _, ok := fields[f]; ok {
			label = strings.Join(strings.Fields(strings.ToLower(row[fields[f]])), " ")
			if label == "" {
				label = strconv.Itoa(k)
			}
		}
		if v, ok := pk.values[label]; ok {
			return nil, fmt.Errorf("on row %d: field %q: label %q already used (by key %d)", ln, f, label, v)
		}
		pk.values[label] = k
		pk.labels[k] = label

		f = "gray"
		if _, ok := fields[f]; !ok {
			continue
		}
		gray, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf(" on row %d: field %q: %v", ln, f, err)
		}
		if gray > 255 {
			return nil, fmt.Errorf(" on row %d: field %q: invalid value %d", ln, f, gray)
		}

		pk.gray[k] = uint8(gray)
	}
	return pk, nil
}

// TSV writes a file with the keys for the pixel colors and labels
// in a time pixelation.
func (pk *PixKey) TSV(w io.Writer) error {
	if len(pk.labels) == 0 {
		for v := range pk.color {
			pk.labels[v] = strconv.Itoa(v)
		}
	}

	tab := csv.NewWriter(w)
	tab.Comma = '\t'
	tab.UseCRLF = true

	head := []string{
		"key",
		"label",
		"color",
	}
	if len(pk.gray) > 0 {
		head = append(head, "gray")
	}
	if err := tab.Write(head); err != nil {
		return err
	}

	pv := make([]int, 0, len(pk.color))
	for v := range pk.color {
		pv = append(pv, v)
	}
	slices.Sort(pv)

	for _, v := range pv {
		c, ok := pk.color[v]
		if !ok {
			continue
		}
		r, g, b, _ := c.RGBA()
		cc := fmt.Sprintf("%d, %d, %d", uint8(r), uint8(g), uint8(b))

		row := []string{
			strconv.Itoa(v),
			pk.labels[v],
			cc,
		}

		if len(pk.gray) > 0 {
			c, ok := pk.gray[v]
			if !ok {
				continue
			}
			row = append(row, strconv.Itoa(int(c)))
		}

		if err := tab.Write(row); err != nil {
			return err
		}
	}

	tab.Flush()
	if err := tab.Error(); err != nil {
		return err
	}
	return nil
}
