// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package pixprob_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/js-arias/earth/stat/pixprob"
)

func TestReadPixel(t *testing.T) {
	data := `
key	prior	comment
0	0.000000	deep ocean
1	0.010000	oceanic plateaus
2	0.050000	continental shelf
3	0.950000	lowlands
4	1.000000	highlands
5	0.001000	ice sheets
`

	want := pixprob.Pixel{
		0: 0,
		1: 0.01,
		2: 0.05,
		3: 0.95,
		4: 1,
		5: 0.001,
	}

	p, err := pixprob.ReadTSV(strings.NewReader(data))
	if err != nil {
		t.Fatalf("unable to read data: %v", err)
	}
	if !reflect.DeepEqual(p, want) {
		t.Errorf("got %v, want %v", p, want)
	}

	vs := []int{0, 1, 2, 3, 4, 5}
	if g := p.Values(); !reflect.DeepEqual(g, vs) {
		t.Errorf("values: got %v, want %v", g, vs)
	}

	for _, v := range vs {
		if p.Prior(v) != want[v] {
			t.Errorf("prior: value %d: got %.6f, want %.6f", v, p.Prior(v), want[v])
		}
	}
}

func TestSet(t *testing.T) {
	p := pixprob.New()
	p.Set(1, 0.01)
	p.Set(2, 0.05)
	p.Set(3, 1.00)

	want := pixprob.Pixel{
		0: 0,
		1: 0.01,
		2: 0.05,
		3: 1.00,
	}
	if !reflect.DeepEqual(p, want) {
		t.Errorf("got %v, want %v", p, want)
	}
	vs := []int{0, 1, 2, 3}
	if g := p.Values(); !reflect.DeepEqual(g, vs) {
		t.Errorf("values: got %v, want %v", g, vs)
	}

	for _, v := range vs {
		if p.Prior(v) != want[v] {
			t.Errorf("prior: value %d: got %.6f, want %.6f", v, p.Prior(v), want[v])
		}
	}
}
