// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package rotation implements a plate tectonic rotation model.
package rotation

import (
	"bufio"
	"cmp"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/js-arias/earth"
	"gonum.org/v1/gonum/num/quat"
	"gonum.org/v1/gonum/spatial/r3"
)

// MillionYears is used to transform rotation ages
// (a float in million years)
// to an integer in years.
const millionYears = 1_000_000

// A Rotation is a rotation model.
type Rotation struct {
	p map[int]*plate
}

// Read decodes a rotation file to produce a set of plates
// each one with its set of rotations.
//
// The rotation file uses the same format used in [GPlates] software
// that are the standard files for tectonic rotations.
// In a rotation file,
// each column is separated by one or more spaces
// and each row represent an Euler rotation angle.
// Fields are:
//
//   - The first column is the ID of the moving plate.
//   - The second column is the the most recent time,
//     in million years.
//   - The third column is the latitude of the Euler pole.
//   - The fourth column is the longitude of the Euler pole.
//   - The fifth column is the angle of the rotation
//     in degrees.
//   - The sixth column is the fixed plate.
//   - Any additional columns are taken as commentaries.
//
// Here is an example of a rotation file:
//
//	1 0.0 90.0 0.0 0.0 0
//	1 37.0 68.0 129.9   7.8 0
//	1 48.0 50.8 142.8   9.8 0
//	1 53.0 40.0 145.0  11.4 0
//	1 83.0 70.5 150.1  20.3 0
//	2  0.0  0.0   0.0   0.0 1
//	2 37.0 70.5 -18.7 -10.4 1
//	2 66.0 80.8  -8.6 -22.5 1
//	2 71.0 80.4 -12.5 -23.9 1
//
// Because old programs use plate ID 999 as comment,
// that plate ID will be ignored.
// Plate ID 0 is interpreted as the Earth rotation axis.
//
// It is important to remember that the rotation file
// is a file with total rotations
// because each rotation is anchored in the present day.
//
// [GPlates]: https://www.gplates.org
func Read(r io.Reader) (Rotation, error) {
	rots := make(map[int]*plate)
	bw := bufio.NewReader(r)
	for i := 1; ; i++ {
		ln, err := bw.ReadString('\n')
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return Rotation{}, fmt.Errorf("row %d", i)
		}

		cols := strings.Fields(ln)
		if len(cols) < 6 {
			continue
		}

		// First column:
		// moving plate
		id, err := strconv.Atoi(cols[0])
		if err != nil {
			return Rotation{}, fmt.Errorf("row %d [ID: %d]: column 'moving plate': %v", i, id, err)
		}

		// ignore plate ID 999
		if id == 999 {
			continue
		}

		p, ok := rots[id]
		if !ok {
			p = &plate{id: id}
			rots[id] = p
		}

		// Second column:
		// time in million years
		t, err := strconv.ParseFloat(cols[1], 64)
		if err != nil {
			return Rotation{}, fmt.Errorf("row %d [ID: %d]: column 'time': %v", i, id, err)
		}

		// Third column:
		// latitude
		lat, err := strconv.ParseFloat(cols[2], 64)
		if err != nil {
			return Rotation{}, fmt.Errorf("row %d [ID: %d]: column 'latitude': %v", i, id, err)
		}
		if lat < -90 || lat > 90 {
			return Rotation{}, fmt.Errorf("row %d [ID: %d]: column 'latitude': bad value %.3f", i, id, lat)
		}

		// Fourth column:
		// longitude
		lon, err := strconv.ParseFloat(cols[3], 64)
		if err != nil {
			return Rotation{}, fmt.Errorf("row %d [ID: %d]: column 'longitude': %v", i, id, err)
		}
		if lat < -180 || lat > 180 {
			return Rotation{}, fmt.Errorf("row %d [ID: %d]: column 'longitude': bad value %.3f", i, id, lon)
		}

		// Fifth column:
		// rotation angle
		ang, err := strconv.ParseFloat(cols[4], 64)
		if err != nil {
			return Rotation{}, fmt.Errorf("row %d [ID: %d]: column 'angle': %v", i, id, err)
		}

		// Sixth column:
		// fixed plate
		fix, err := strconv.Atoi(cols[5])
		if err != nil {
			return Rotation{}, fmt.Errorf("row %d [ID: %d]: column 'fixed plate': %v", i, id, err)
		}

		rot := Euler{
			T:     int64(t * millionYears),
			E:     earth.NewPoint(lat, lon),
			Angle: earth.ToRad(ang),
			Fix:   fix,
		}

		// check if the rotation is repeated
		rep := false
		for _, r := range p.rot {
			if r.T == rot.T && r.Fix == rot.Fix {
				rep = true
				break
			}
		}
		if rep {
			continue
		}

		p.rot = append(p.rot, rot)
	}

	for _, p := range rots {
		slices.SortFunc(p.rot, func(a, b Euler) int {
			return cmp.Compare(a.T, b.T)
		})

		// add a zero rotation by default
		// if not defined.
		if p.rot[0].T > 0 {
			r := Euler{
				E:   earth.NorthPole,
				Fix: p.rot[0].Fix,
			}
			p.rot = append([]Euler{r}, p.rot...)
		}

		// remove wrong jumps
		for i := 0; i < len(p.rot); i++ {
			if i == 0 {
				continue
			}
			if i+1 == len(p.rot) {
				continue
			}

			r := p.rot[i]

			j := i + 1
			k := i - 1

			if p.rot[j].T != r.T && p.rot[k].T != r.T {
				continue
			}

			prev := -1
			for x := i; x >= 0; x-- {
				if p.rot[x].T != r.T {
					prev = p.rot[x].Fix
					break
				}
			}
			if prev == p.rot[i].Fix {
				continue
			}
			post := -1
			for x := i; x < len(p.rot); x++ {
				if p.rot[x].T != r.T {
					post = p.rot[x].Fix
					break
				}
			}
			if post == p.rot[i].Fix {
				continue
			}

			p.rot = slices.Delete(p.rot, i, i+1)
		}

		// check that conjugate
		// (or fixed)
		// plate jumps are well sorted:
		// so any time stage
		// will be bounded by two rotations
		// relative to the same fixed plate.
		for i, r := range p.rot {
			if i == 0 {
				continue
			}
			if i+1 == len(p.rot) {
				continue
			}
			j := i + 1
			if p.rot[j].T != r.T {
				continue
			}
			k := i - 1
			if p.rot[k].Fix != r.Fix {
				p.rot[i], p.rot[j] = p.rot[j], p.rot[i]
			}
		}
	}
	return Rotation{rots}, nil
}

// Rotation returns a total rotation
// (i.e. a rotation from current time)
// for a plate at a particular time
// (in years).
// It returns false if there are no rotation defined
// at the indicated time.
//
// The rotations produced by this function
// are based on the descriptions in
// Cox. A, Hart R.B. 1986.
// Plate tectonics: How it works.
// Blackwell, Palo Alto (USA),
// in particular chapter 7,
// but using quaternions instead of rotation matrices.
func (r Rotation) Rotation(plate int, t int64) (r3.Rotation, bool) {
	p, ok := r.p[plate]
	if !ok {
		return r3.Rotation{}, false
	}
	if len(p.rot) == 0 {
		return r3.Rotation{}, false
	}

	// Make the global circuit
	// to calculate the rotation
	var qt quat.Number
	for i := 0; ; i++ {
		x := p.timePos(t)
		if x == -1 {
			return r3.Rotation{}, false
		}

		tot := quat.Number(r3.NewRotation(p.rot[x].Angle, p.rot[x].E.Vector()))
		if p.rot[x].T != t {
			stage := p.stage(x, t)
			tot = quat.Mul(stage, tot)
		}
		if i == 0 {
			qt = tot
		} else {
			qt = quat.Mul(tot, qt)
		}

		p, ok = r.p[p.rot[x].Fix]
		if !ok {
			break
		}
	}

	return r3.Rotation(qt), true
}

// Euler returns the list of Euler rotations
// for a given plate.
func (r Rotation) Euler(plate int) []Euler {
	p, ok := r.p[plate]
	if !ok {
		return nil
	}

	e := make([]Euler, len(p.rot))
	copy(e, p.rot)
	return e
}

// Plates return the plates defined for a rotation model.
func (r Rotation) Plates() []int {
	plates := make([]int, 0, len(r.p))
	for id := range r.p {
		plates = append(plates, id)
	}
	slices.Sort(plates)
	return plates
}

// A Plate is a collection of rotations
// for the indicated plate.
type plate struct {
	id  int     // ID of the plate
	rot []Euler // rotations
}

// Stage returns the stage rotation between two total rotations,
// and scale it to the time we are looking for
// (follows the procedure given by Cox & Hart, pp. 245-246).
func (p *plate) stage(x int, t int64) quat.Number {
	q1 := quat.Number(r3.NewRotation(-p.rot[x].Angle, p.rot[x].E.Vector()))
	q2 := quat.Number(r3.NewRotation(p.rot[x-1].Angle, p.rot[x-1].E.Vector()))
	s := quat.Mul(q2, q1)

	delta := float64(p.rot[x].T-t) / float64(p.rot[x].T-p.rot[x-1].T)

	// In quaternions,
	// exponential to delta
	// give us the fraction of the rotation.
	return quat.Pow(s, quat.Number{Real: delta})
}

// TimePos returns the position of the time
// that adjust better to the required rotation.
func (p *plate) timePos(t int64) int {
	for i, v := range p.rot {
		if v.T >= t {
			return i
		}
	}
	return -1
}

// Euler is a rotation of a moving plate
// relative to a fixed plate.
type Euler struct {
	T     int64       // starting time for the rotation (in years)
	E     earth.Point // Euler pole
	Angle float64     // angle of the rotation in radians
	Fix   int         // ID of the fixed plate
}

// Rotate returns a vector
// from a given coordinate
// rotated using the indicated rotation.
func Rotate(r r3.Rotation, lat, lon float64) r3.Vec {
	return r.Rotate(earth.NewPoint(lat, lon).Vector())
}

// Inverse returns the inverse of a rotation.
func Inverse(r r3.Rotation) r3.Rotation {
	return r3.Rotation(quat.Conj(quat.Number(r)))
}
