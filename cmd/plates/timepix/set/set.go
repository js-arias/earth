// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package set implements a command to set pixel values
// to a time pixelation.
package set

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/js-arias/command"
	"github.com/js-arias/earth"
	"github.com/js-arias/earth/model"
	"golang.org/x/exp/slices"
)

var Command = &command.Command{
	Usage: `set [--from <age>] [--to <age>] [--at <age>]
	--in <time-pix-file> <time-pix-file>`,
	Short: "set pixels of a time pixelation",
	Long: `
Command set reads pixels from time pixelation file, and set that values into a
time pixelation.

The flag --in is required and is used to provide the name of the input file.
The input file is a time pixelation file, and all pixels in the time frame
will be set to the indicated values. The file is a TSV file with the following
columns:
	- equator, for the number of pixels at the equator
	- age, the age of the time stage (in years)
	- stage-pixel, the pixel ID at the time stage
	- value, an integer value

Here is an example file:

	equator	age	stage-pixel	value
	360	100000000	19051	1
	360	100000000	19055	2
	360	140000000	20051	1
	360	140000000	20055	2

If a pixel has a value of 0, then it will be deleted from the time pixelation.

The argument of the command is the file that contains the time pixelation.
This argument is required.

By default, all time stages of the source pixels (as defined by --in) will be
set. With the flags --from and --to, it will use only the stages inside of the
indicated ages (in million years). Another possibility is using the flag --at
to set a particular time stage.
	`,
	SetFlags: setFlags,
	Run:      run,
}

var inFlag string
var fromFlag float64
var toFlag float64
var atFlag float64

func setFlags(c *command.Command) {
	c.Flags().Float64Var(&fromFlag, "from", -1, "")
	c.Flags().Float64Var(&toFlag, "to", -1, "")
	c.Flags().Float64Var(&atFlag, "at", -1, "")
	c.Flags().StringVar(&inFlag, "in", "", "")
}

// MillionYears is used to transform ages in the flags
// (floats in million years)
// to an integer in years.
const millionYears = 1_000_000

func run(c *command.Command, args []string) error {
	if len(args) < 1 {
		return c.UsageError("expecting time pixelation file")
	}

	if inFlag == "" {
		return c.UsageError("flag --in must be set")
	}
	output := args[0]

	tp, err := readTimePix(output)
	if err != nil {
		return err
	}

	rows, err := readSourcePixels(inFlag, tp.Pixelation())
	if err != nil {
		return nil
	}
	if len(rows) == 0 {
		return nil
	}

	var stages []int64
	if atFlag >= 0 {
		stages = []int64{tp.CloserStageAge(int64(atFlag * millionYears))}
	} else {
		st := tp.Stages()
		from := st[len(st)-1]
		if fromFlag >= 0 {
			from = int64(fromFlag * millionYears)
		}
		to := st[0]
		if toFlag >= 0 {
			to = int64(toFlag * millionYears)
		}
		stages = make([]int64, 0, len(st))
		for _, a := range st {
			if a > from {
				continue
			}
			if a < to {
				continue
			}
			stages = append(stages, a)
		}
		if len(stages) == 0 {
			return nil
		}
		slices.Sort(stages)
	}

	setTimeValue(tp, rows, stages)

	if err := writeTimePix(output, tp); err != nil {
		return err
	}
	return nil
}

func setTimeValue(tp *model.TimePix, rows []pixRow, ages []int64) {
	for _, r := range rows {
		if _, ok := slices.BinarySearch(ages, r.age); !ok {
			continue
		}
		if r.val == 0 {
			tp.Del(r.age, r.pix)
			continue
		}
		tp.Set(r.age, r.pix, r.val)
	}
}

func readTimePix(name string) (*model.TimePix, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tp, err := model.ReadTimePix(f, nil)
	if err != nil {
		return nil, fmt.Errorf("when reading file %q: %v", name, err)
	}
	return tp, nil
}

type pixRow struct {
	age int64
	pix int
	val int
}

var tpHeader = []string{
	"equator",
	"age",
	"stage-pixel",
	"value",
}

func readSourcePixels(name string, pix *earth.Pixelation) ([]pixRow, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := readRows(f, pix)
	if err != nil {
		return nil, fmt.Errorf("on file %q: %v", name, err)
	}
	return rows, nil
}

func readRows(r io.Reader, pix *earth.Pixelation) ([]pixRow, error) {
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
	for _, h := range tpHeader {
		if _, ok := fields[h]; !ok {
			return nil, fmt.Errorf("expecting field %q", h)
		}
	}

	var rows []pixRow
	for {
		row, err := tab.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		ln, _ := tab.FieldPos(0)
		if err != nil {
			return nil, fmt.Errorf("on row %d: %v", ln, err)
		}

		f := "equator"
		eq, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		if eq != pix.Equator() {
			return nil, fmt.Errorf("on row %d: field %q: got %d pixels, want %d", ln, f, eq, pix.Equator())
		}

		f = "age"
		age, err := strconv.ParseInt(row[fields[f]], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}

		f = "stage-pixel"
		px, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}
		if px >= pix.Len() {
			return nil, fmt.Errorf("on row %d: field %q: invalid pixel value %d", ln, f, px)
		}

		f = "value"
		v, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf("on row %d: field %q: %v", ln, f, err)
		}

		rows = append(rows, pixRow{
			age: age,
			pix: px,
			val: v,
		})
	}

	return rows, nil
}

func writeTimePix(name string, tp *model.TimePix) (err error) {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer func() {
		e := f.Close()
		if e != nil && err == nil {
			err = e
		}
	}()

	if err := tp.TSV(f); err != nil {
		return err
	}
	return nil
}
