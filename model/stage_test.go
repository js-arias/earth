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

func TestReadStageRot(t *testing.T) {
	data := makeRecons(t)

	var buf bytes.Buffer
	if err := data.TSV(&buf); err != nil {
		t.Fatalf("while writing data: %v", err)
	}

	stg, err := model.ReadStageRot(strings.NewReader(buf.String()), nil)
	if err != nil {
		t.Fatalf("while reading data: %v", err)
	}
	testStageRot(t, stg)
}

func testStageRot(t testing.TB, stg *model.StageRot) {
	t.Helper()

	stages := []int64{100_000_000, 140_000_000}
	if st := stg.Stages(); !reflect.DeepEqual(st, stages) {
		t.Errorf("stages: got %v, want %v", st, stages)
	}

	pix100to140 := map[int][]int{
		19051: {20051},
		19055: {20055, 20056},
		19409: {20409},
		19766: {20766},
		20122: {21122},
		20479: {21479},
		20480: {21479},
	}

	pix140to100 := map[int][]int{
		20051: {19051},
		20055: {19055},
		20056: {19055},
		20409: {19409},
		20766: {19766},
		21122: {20122},
		21479: {20479, 20480},
	}

	y2o := stg.YoungToOld(100_000_000)
	if !reflect.DeepEqual(y2o, pix100to140) {
		t.Errorf("young to old: got %v, want %v", y2o, pix100to140)
	}
	if o := stg.OldAge(100_000_000); o != 140_000_000 {
		t.Errorf("old age: got %d, want %d", o, 140_000_000)
	}

	o2y := stg.OldToYoung(140_000_000)
	if !reflect.DeepEqual(o2y, pix140to100) {
		t.Errorf("old to young: got %v, want %v", o2y, pix140to100)
	}
	if y := stg.YoungAge(140_000_000); y != 100_000_000 {
		t.Errorf("young age: got %d, want %d", y, 100_000_000)
	}

	if c := stg.CloserStageAge(125_000_000); c != 100_000_000 {
		t.Errorf("closer stage age: got %d, want %d", c, 100_000_000)
	}
}
