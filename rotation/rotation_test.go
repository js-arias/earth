// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package rotation_test

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/js-arias/earth"
	"github.com/js-arias/earth/rotation"
	"gonum.org/v1/gonum/spatial/r3"
)

// This test just made a simple rotation.
// It is numerically based on the box 7-3 of Cox & Hart.
func TestSimple(t *testing.T) {
	simple := "1 90.0 0.0 0.0 0\n1 100.0 -37 -48 65 0\n"
	rots, err := rotation.Read(strings.NewReader(simple))
	if err != nil {
		t.Fatalf("when reading rotation: %v", err)
	}
	r, ok := rots.Rotation(1, 100_000_000)
	if !ok {
		t.Fatalf("want rotation at %d\n", 100_000_000)
	}

	// Numerical test
	want := earth.NewPoint(30, 113.2).Vector()
	v := rotation.Rotate(r, 20, 130)
	if isDiff(v, want) {
		t.Errorf("rotation: got %v, want %v", v, want)
	}

	// Inverse rotation
	org := earth.NewPoint(20, 130).Vector()
	g := rotation.Rotate(rotation.Inverse(r), 30, 113.2)
	if isDiff(g, org) {
		t.Errorf("inverse rotation: got %v, want %v", g, org)
	}

	// Test using r3 rotations
	testRotation(t, r, newRotation(65, -37, -48), 20, 130)
}

// This is a test for an intermediate rotation
// between two total reconstruction poles.
// It is based on the example of pag. 246
// and table 7-1 of Cox & Hart.
func TestIntermediate(t *testing.T) {
	in := `1 0.0 90.0 0.0 0.0 0
1 37.0 68.0 129.9 -7.8 0
1 48.0 50.8 142.8 -9.8 0
1 53.0 40.0 145.0 -11.4 0
1 83.0 70.5 150.1 -20.3 0
1 90.0 75.5 152.9 -24.2 0
	`
	rots, err := rotation.Read(strings.NewReader(in))
	if err != nil {
		t.Fatalf("when reading rotations: %v", err)
	}
	r, ok := rots.Rotation(1, 40_000_000)
	if !ok {
		t.Fatalf("want rotation at %d\n", 40_000_000)
	}
	testRotation(t, r, newRotation(8.25, -62.65, -44.39), 20, 130)
}

// Table 7-3 of Cox & Hart.
var coxHartTable73 = `1 0.0 90.0 0.0 0.0 0
1 37.0 68.0 129.9   7.8 0
1 48.0 50.8 142.8   9.8 0
1 53.0 40.0 145.0  11.4 0
1 83.0 70.5 150.1  20.3 0
2  0.0  0.0   0.0   0.0 1
2 37.0 70.5 -18.7 -10.4 1
2 66.0 80.8  -8.6 -22.5 1
2 71.0 80.4 -12.5 -23.9 1
3  0.0  0.0   0.0   0.0 2
3 40.0  5.8 -37.2   7.2 2
3 50.0 12.0 -48.6   7.5 2
3 83.0 19.7 -43.8  19.2 2
4  0.0  0.0   0.0   0.0 3
4 37.0 11.9  34.4 -20.5 3  
4 42.0 10.3  34.8 -23.6 3
4 50.0 11.9  30.8 -30.9 3
5  0.0  0.0   0.0   0.0 4
5 50.0  0.0   0.0   0.0 4
5 63.0  8.9 -26.6  17.2 4
5 83.0  5.6  -4.7  38.6 4
`

// This test is for a rotation
// in a rotation  hierarchy
// (a global circuit in Cox & Hart).
// It is based on the example of pp. 248-251
// and table 7-3 of Cox & Hart.
func TestCircuit(t *testing.T) {
	rots, err := rotation.Read(strings.NewReader(coxHartTable73))
	if err != nil {
		t.Fatalf("when reading rotations: %v", err)
	}
	r, ok := rots.Rotation(5, 40_000_000)
	if !ok {
		t.Fatalf("want rotation at %d\n", 40_000_000)
	}
	testRotation(t, r, newRotation(-24.34, 17.21, 34.89), 20, 130)
}

func TestUnordered(t *testing.T) {
	in := `
5 83.0  5.6  -4.7  38.6 4
5 63.0  8.9 -26.6  17.2 4
5 50.0  0.0   0.0   0.0 4
5  0.0  0.0   0.0   0.0 4
4 50.0 11.9  30.8 -30.9 3
4 42.0 10.3  34.8 -23.6 3
4 37.0 11.9  34.4 -20.5 3  
4  0.0  0.0   0.0   0.0 3
3 83.0 19.7 -43.8  19.2 2
3 50.0 12.0 -48.6   7.5 2
3 40.0  5.8 -37.2   7.2 2
3  0.0  0.0   0.0   0.0 2
2 71.0 80.4 -12.5 -23.9 1
2 66.0 80.8  -8.6 -22.5 1
2 37.0 70.5 -18.7 -10.4 1
2  0.0  0.0   0.0   0.0 1
1 83.0 70.5 150.1  20.3 0
1 53.0 40.0 145.0  11.4 0
1 48.0 50.8 142.8   9.8 0
1 37.0 68.0 129.9   7.8 0
1 0.0 90.0 0.0 0.0 0
`
	rots, err := rotation.Read(strings.NewReader(in))
	if err != nil {
		t.Fatalf("when reading rotations: %v", err)
	}
	r, ok := rots.Rotation(5, 40_000_000)
	if !ok {
		t.Fatalf("want rotation at %d\n", 40_000_000)
	}
	testRotation(t, r, newRotation(-24.34, 17.21, 34.89), 20, 130)
}

func TestJump(t *testing.T) {
	in := `
5 83.0  5.6  -4.7  38.6 4
5 63.0  8.9 -26.6  17.2 4
5 50.0  0.0   0.0   0.0 4
5  0.0  0.0   0.0   0.0 4
4 83.0 70.5 150.1  20.3 1
4 50.0 68.0 129.9   7.8 1
4 50.0 11.9  30.8 -30.9 3
4 42.0 10.3  34.8 -23.6 3
4 37.0 11.9  34.4 -20.5 3  
4 37.0 70.5 150.1  20.3 2
4  0.0  0.0   0.0   0.0 2
3 83.0 19.7 -43.8  19.2 2
3 50.0 12.0 -48.6   7.5 2
3 40.0  5.8 -37.2   7.2 2
3  0.0  0.0   0.0   0.0 2
2 71.0 80.4 -12.5 -23.9 1
2 66.0 80.8  -8.6 -22.5 1
2 37.0 70.5 -18.7 -10.4 1
2  0.0  0.0   0.0   0.0 1
1 83.0 70.5 150.1  20.3 0
1 53.0 40.0 145.0  11.4 0
1 48.0 50.8 142.8   9.8 0
1 37.0 68.0 129.9   7.8 0
1 0.0 90.0 0.0 0.0 0
`

	rots, err := rotation.Read(strings.NewReader(in))
	if err != nil {
		t.Fatalf("when reading rotations: %v", err)
	}
	r, ok := rots.Rotation(5, 40_000_000)
	if !ok {
		t.Fatalf("want rotation at %d\n", 40_000_000)
	}
	testRotation(t, r, newRotation(-24.34, 17.21, 34.89), 20, 130)
}

func TestRepeated(t *testing.T) {
	in := `1 0.0 90.0 0.0 0.0 0
1 37.0 68.0 129.9 -7.8 0
1 48.0 50.8 142.8 -9.8 0
1 48.0 50.8 142.8 -9.8 0
1 48.0 50.8 142.8 -9.8 0
1 48.0 50.8 142.8 -9.8 0
1 53.0 40.0 145.0 -11.4 0
1 83.0 70.5 150.1 -20.3 0
1 90.0 75.5 152.9 -24.2 0
	`
	rots, err := rotation.Read(strings.NewReader(in))
	if err != nil {
		t.Fatalf("when reading rotations: %v", err)
	}
	r, ok := rots.Rotation(1, 40_000_000)
	if !ok {
		t.Fatalf("want rotation at %d\n", 40_000_000)
	}
	if math.IsNaN(r.Imag) {
		t.Errorf("nan rotation: %v", r)
	}
}

func TestPlates(t *testing.T) {
	rots, err := rotation.Read(strings.NewReader(coxHartTable73))
	if err != nil {
		t.Fatalf("when reading rotations: %v", err)
	}

	plates := rots.Plates()
	want := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(plates, want) {
		t.Errorf("plates: got %v, want %v", plates, want)
	}
}

func TestEuler(t *testing.T) {
	rots, err := rotation.Read(strings.NewReader(coxHartTable73))
	if err != nil {
		t.Fatalf("when reading rotations: %v", err)
	}

	want := []rotation.Euler{
		{E: earth.NewPoint(0, 0), Fix: 1},
		{T: 37_000_000, E: earth.NewPoint(70.5, -18.7), Angle: earth.ToRad(-10.4), Fix: 1},
		{T: 66_000_000, E: earth.NewPoint(80.8, -8.6), Angle: earth.ToRad(-22.5), Fix: 1},
		{T: 71_000_000, E: earth.NewPoint(80.4, -12.5), Angle: earth.ToRad(-23.9), Fix: 1},
	}
	e := rots.Euler(2)
	if !reflect.DeepEqual(e, want) {
		t.Errorf("euler: got %v, want %v", e, want)
	}
}

func TestMultiJump(t *testing.T) {
	in := `505  0.0   0.0    0.0    0.0  501 !!
505 50.0 -28.83 -123.27   40.16  501 !!
505 65.0 -33.6 -123.6   75.56  501 !!
505 65.0 -17.21 -138.31  116.59  000 !! crs 04/24/98
505 65.0 -22.55 -127.64  106.34  503 !! crs 04/24/98
505 96.0 -22.55 -127.64  106.34  503 !!
	`

	rots, err := rotation.Read(strings.NewReader(in))
	if err != nil {
		t.Fatalf("when reading rotations: %v", err)
	}
	want := []rotation.Euler{
		{E: earth.NewPoint(0, 0), Fix: 501},
		{T: 50_000_000, E: earth.NewPoint(-28.83, -123.27), Angle: earth.ToRad(40.16), Fix: 501},
		{T: 65_000_000, E: earth.NewPoint(-33.6, -123.6), Angle: earth.ToRad(75.56), Fix: 501},
		{T: 65_000_000, E: earth.NewPoint(-22.55, -127.64), Angle: earth.ToRad(106.34), Fix: 503},
		{T: 96_000_000, E: earth.NewPoint(-22.55, -127.64), Angle: earth.ToRad(106.34), Fix: 503},
	}

	e := rots.Euler(505)
	if !reflect.DeepEqual(e, want) {
		t.Errorf("euler: got %v, want %v", e, want)
	}
}

func testRotation(t testing.TB, r, rot r3.Rotation, lat, lon float64) {
	t.Helper()

	v := rotation.Rotate(r, lat, lon)
	want := rot.Rotate(earth.NewPoint(lat, lon).Vector())

	if isDiff(v, want) {
		t.Errorf("rotation: got %v, want %v", v, want)
	}

	org := earth.NewPoint(lat, lon).Vector()
	got := rotation.Inverse(r).Rotate(want)
	if isDiff(got, org) {
		t.Errorf("inverse rotation: got %v, want %v", got, org)
	}
}

func newRotation(euler, lat, lon float64) r3.Rotation {
	return r3.NewRotation(earth.ToRad(euler), earth.NewPoint(lat, lon).Vector())
}

func isDiff(v, w r3.Vec) bool {
	max := 0.001
	if math.Abs(v.X-w.X) > max {
		return true
	}
	if math.Abs(v.Y-w.Y) > max {
		return true
	}
	if math.Abs(v.Z-w.Z) > max {
		return true
	}
	return false
}
