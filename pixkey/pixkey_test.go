// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package pixkey_test

import (
	"bytes"
	"fmt"
	"image/color"
	"reflect"
	"strings"
	"testing"

	"github.com/js-arias/earth/pixkey"
)

func TestPixKey(t *testing.T) {
	pk := pixkey.New()

	pk.SetColor(color.RGBA{54, 75, 154, 255}, 0)
	pk.SetColor(color.RGBA{74, 123, 154, 255}, 1)
	pk.SetColor(color.RGBA{254, 218, 139, 255}, 3)

	pk.SetGray(color.RGBA{255, 255, 255, 255}, 0)
	pk.SetGray(color.RGBA{235, 235, 235, 255}, 1)
	pk.SetGray(color.RGBA{195, 195, 195, 255}, 3)

	pk.SetLabel(0, "ocean")
	pk.SetLabel(1, "oceanic plateaus")
	pk.SetLabel(3, "lands")

	testPixKey(t, pk)

	var buf bytes.Buffer
	if err := pk.TSV(&buf); err != nil {
		t.Fatalf("while writing data: %v", err)
	}

	np, err := pixkey.Read(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("while reading data: %v", err)
	}

	testPixKey(t, np)
}

func testPixKey(t testing.TB, pk *pixkey.PixKey) {
	t.Helper()

	colors := map[int]color.RGBA{
		0: {54, 75, 154, 255},
		1: {74, 123, 154, 255},
		3: {254, 218, 139, 255},
	}
	for v, c := range colors {
		w := colorString(c)
		gc, ok := pk.Color(v)
		if !ok {
			t.Errorf("color: value %d: got <nil> want %q", v, w)
			continue
		}
		g := colorString(gc)
		if g != w {
			t.Errorf("color: value %d: got %q, want %q", v, g, w)
		}
	}

	grays := map[int]uint8{
		0: 255,
		1: 235,
		3: 195,
	}

	for v, c := range grays {
		gc, ok := pk.Gray(v)
		if !ok {
			t.Errorf("gray: value %d: got <nil>, want '%d'", v, c)
			continue
		}
		r, _, _, _ := gc.RGBA()
		if uint8(r>>8) != c {
			t.Errorf("gray: value %d: got '%d', want '%d'", v, uint8(r), c)
		}
	}

	keys := []int{0, 1, 3}
	got := pk.Keys()
	if !reflect.DeepEqual(got, keys) {
		t.Errorf("keys: got %v, want %v", got, keys)
	}

	labels := map[int]string{
		0: "ocean",
		1: "oceanic plateaus",
		3: "lands",
	}
	for v, l := range labels {
		g := pk.Label(v)
		if g != l {
			t.Errorf("labels: value %d: got %q, want %q", v, g, l)
		}

		gv := pk.Key(l)
		if gv != v {
			t.Errorf("key: label %q: got %d, want %d", l, gv, v)
		}
	}
}

func colorString(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("%d,%d,%d", uint8(r), uint8(g), uint8(b))
}
