// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package model_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/js-arias/earth"
	"github.com/js-arias/earth/model"
)

func TestTimePix(t *testing.T) {
	data := makeRecons(t)
	tot := model.NewTotal(data)

	tp := model.NewTimePix(tot.Pixelation())
	setStage(tp, tot, 100_000_000)
	setStage(tp, tot, 140_000_000)

	testTimePix(t, tp)

	var buf bytes.Buffer
	if err := tp.TSV(&buf); err != nil {
		t.Fatalf("while writing data: %v", err)
	}

	np, err := model.ReadTimePix(strings.NewReader(buf.String()), nil)
	if err != nil {
		t.Fatalf("while reading data: %v", err)
	}

	testTimePix(t, np)
}

func TestTimePixDelete(t *testing.T) {
	data := makeRecons(t)
	tot := model.NewTotal(data)

	tp := model.NewTimePix(tot.Pixelation())
	setStage(tp, tot, 100_000_000)
	setStage(tp, tot, 140_000_000)

	age := int64(100_000_000)

	tp.Del(age, 19409)
	tp.Del(age, 8000)

	vals100 := map[int]int{
		15000: 0,
		19051: 1,
		19055: 1,
		19409: 0,
		19766: 1,
		20122: 1,
		20479: 1,
		20480: 1,
	}

	for id, x := range vals100 {
		v, ok := tp.At(age, id)
		if !ok {
			t.Errorf("time %d: pixel %d: not found, want %d", age, id, x)
		}
		if v != x {
			t.Errorf("time %d: pixel %d: got %d, want %d", age, id, v, x)
		}
	}

	st100 := map[int]int{
		19051: 1,
		19055: 1,
		19766: 1,
		20122: 1,
		20479: 1,
		20480: 1,
	}
	if st := tp.Stage(age); !reflect.DeepEqual(st, st100) {
		t.Errorf("stage at 100_000_000: got %v, want %v", st, st100)
	}
}

func setStage(tp *model.TimePix, tot *model.Total, age int64) {
	st := tot.Rotation(age)
	for _, ids := range st {
		for _, id := range ids {
			tp.Set(age, id, 1)
		}
	}
}

func testTimePix(t testing.TB, tp *model.TimePix) {
	t.Helper()

	if eq := tp.Pixelation().Equator(); eq != 360 {
		t.Errorf("pixelation: got %d pixels at equator, want %d", eq, 360)
	}

	stages := []int64{100_000_000, 140_000_000}
	if st := tp.Stages(); !reflect.DeepEqual(st, stages) {
		t.Errorf("stages: got %v, want %v", st, stages)
	}
	if o, y := tp.Bounds(100_000_000); y < 100_000_000 || o > 140_000_000 {
		t.Errorf("bounds: at %d, got %d-%d, want %d-%d", 100_000_000, o, y, 140_000_000, 100_000_000)
	}
	if o, y := tp.Bounds(125_000_000); y < 100_000_000 || o > 140_000_000 {
		t.Errorf("bounds: at %d, got %d-%d, want %d-%d", 125_000_000, o, y, 140_000_000, 100_000_000)
	}
	if o, y := tp.Bounds(145_000_000); y < 140_000_000 || o > earth.Age {
		t.Errorf("bounds: at %d, got %d-%d, want %d-%d", 145_000_000, o, y, earth.Age, 140_000_000)
	}

	vals100 := map[int]int{
		15000: 0,
		19051: 1,
		19055: 1,
		19409: 1,
		19766: 1,
		20122: 1,
		20479: 1,
		20480: 1,
	}
	vals140 := map[int]int{
		15000: 0,
		20051: 1,
		20055: 1,
		20056: 1,
		20409: 1,
		20766: 1,
		21122: 1,
		21479: 1,
	}

	age := int64(100_000_000)
	for id, x := range vals100 {
		v, ok := tp.At(age, id)
		if !ok {
			t.Errorf("time %d: pixel %d: not found, want %d", age, id, x)
		}
		if v != x {
			t.Errorf("time %d: pixel %d: got %d, want %d", age, id, v, x)
		}
	}

	age = 140_000_000
	for id, x := range vals140 {
		v, ok := tp.At(age, id)
		if !ok {
			t.Errorf("time %d: pixel %d: not found, want %d", age, id, x)
		}
		if v != x {
			t.Errorf("time %d: pixel %d: got %d, want %d", age, id, v, x)
		}
	}

	age = 150_000_000
	for id := range vals140 {
		if _, ok := tp.At(age, id); ok {
			t.Errorf("time %d: pixel %d found", age, id)
		}
	}

	if a := tp.ClosestStageAge(150_000_000); a != 140_000_000 {
		t.Errorf("closest stage at 150_000_000: got %d, want %d", a, 140_000_000)
	}

	for id, x := range vals140 {
		v := tp.AtClosest(age, id)
		if v != x {
			t.Errorf("time closest %d (%d): pixel %d: got %d, want %d", age, tp.ClosestStageAge(age), id, v, x)
		}
	}

	st100 := map[int]int{
		19051: 1,
		19055: 1,
		19409: 1,
		19766: 1,
		20122: 1,
		20479: 1,
		20480: 1,
	}
	if st := tp.Stage(100_000_000); !reflect.DeepEqual(st, st100) {
		t.Errorf("stage at 100_000_000: got %v, want %v", st, st100)
	}
}
