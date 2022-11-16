// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package model_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/js-arias/earth/model"
)

func TestReadTotal(t *testing.T) {
	data := makeRecons(t)
	testTotal(t, model.NewTotal(data))

	var buf bytes.Buffer
	if err := data.TSV(&buf); err != nil {
		t.Fatalf("while writing data: %v", err)
	}

	tot, err := model.ReadTotal(strings.NewReader(buf.String()), nil)
	if err != nil {
		t.Fatalf("while reading data: %v", err)
	}

	testTotal(t, tot)
}

func testTotal(t testing.TB, tot *model.Total) {
	t.Helper()

	if eq := tot.Pixelation().Equator(); eq != 360 {
		t.Errorf("pixelation: got %v pixels at equator, want %d", eq, 360)
	}

	stages := []int64{100_000_000, 140_000_000}
	if st := tot.Stages(); !reflect.DeepEqual(st, stages) {
		t.Errorf("stages: got %v, want %v", st, stages)
	}

	pix100 := map[int][]int{
		17051: {19051},
		17055: {19055},
		17409: {19409},
		17766: {19766},
		18122: {20122},
		18479: {20479, 20480},
	}
	pix140 := map[int][]int{
		17051: {20051},
		17055: {20055, 20056},
		17409: {20409},
		17766: {20766},
		18122: {21122},
		18479: {21479},
	}

	if l := tot.Rotation(100_000_000); !reflect.DeepEqual(l, pix100) {
		t.Errorf("pixels at stage 100: got %v, want %v", l, pix100)
	}
	if l := tot.Rotation(140_000_000); !reflect.DeepEqual(l, pix140) {
		t.Errorf("pixels at stage 100: got %v, want %v", l, pix140)
	}

	// Ages given to the model might not be exact
	if l := tot.Rotation(110_000_000); !reflect.DeepEqual(l, pix100) {
		t.Errorf("pixels at stage 100: got %v, want %v", l, pix100)
	}
	if l := tot.Rotation(150_000_000); !reflect.DeepEqual(l, pix140) {
		t.Errorf("pixels at stage 100: got %v, want %v", l, pix140)
	}
}
