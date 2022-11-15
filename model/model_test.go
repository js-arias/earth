// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package model_test

import (
	"reflect"
	"testing"

	"github.com/js-arias/earth"
	"github.com/js-arias/earth/model"
)

func TestNewRecon(t *testing.T) {
	r := makeRecons(t)
	testRecons(t, r)
}

func testRecons(t testing.TB, rec *model.Recons) {
	t.Helper()

	plates := []int{59_999}
	if p := rec.Plates(); !reflect.DeepEqual(p, plates) {
		t.Errorf("plates: got %v, want %v", p, plates)
	}

	pixels := []int{17051, 17055, 17409, 17766, 18122, 18479}
	if pix := rec.Pixels(59_999); !reflect.DeepEqual(pix, pixels) {
		t.Errorf("pixels: got %v, want %v", pix, pixels)
	}

	stages := []int64{100_000_000, 140_000_000}
	if st := rec.Stages(); !reflect.DeepEqual(st, stages) {
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
	if ps := rec.PixStage(59_999, 100_000_000); !reflect.DeepEqual(ps, pix100) {
		t.Errorf("pixels at stage 100: got %v, want %v", ps, pix100)
	}
	if ps := rec.PixStage(59_999, 140_000_000); !reflect.DeepEqual(ps, pix140) {
		t.Errorf("pixels at stage 140: got %v, want %v", ps, pix140)
	}
}

func makeRecons(t testing.TB) *model.Recons {
	t.Helper()

	rec := model.NewRecons(earth.NewPixelation(360))

	locs := []struct {
		age int64
		loc map[int][]int
	}{
		{
			age: 100_000_000,
			loc: map[int][]int{
				17051: {19051},
				17055: {19055},
				17409: {19409},
				17766: {19766},
				18122: {20122},
				18479: {20479, 20480},
			},
		},
		{
			age: 140_000_000,
			loc: map[int][]int{
				17051: {20051},
				17055: {20055, 20056},
			},
		},
		{
			age: 140_000_000,
			loc: map[int][]int{
				17051: {20051},
				17055: {20055},
				17409: {20409},
				17766: {20766},
				18122: {21122},
				18479: {21479},
			},
		},
	}

	for _, l := range locs {
		rec.Add(59_999, l.loc, l.age)
	}

	return rec
}
