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
	"slices"
	"strconv"
	"strings"

	"github.com/js-arias/command"
	"github.com/js-arias/earth"
	"github.com/js-arias/earth/model"
)

var Command = &command.Command{
	Usage: `set [--from <age>] [--to <age>] [--at <age>]
	[-f|--format <format>] --in <model-file> <time-pix-file>`,
	Short: "set pixels of a time pixelation",
	Long: `
Command set reads pixels from time pixelation file, and set that values into a
time pixelation.

The flag --in is required and is used to provide the name of the input file.
By default, the input file is a time pixelation file, other kind of files can
be defined using the flag --format, or -f. Valid format are:

	latlon	 a location file
	timepix  default value, a time pixelation file

All pixels defined in the input file, and inside the indicated time frame will
be to the indicated values. If a pixel has a value of 0, then it will be
deleted from the time pixelation.

The locations file is a tab-delimited text file with the following columns:
	
	- age        the age for the value in years
	- latitude   the geographic latitude of the location,
	             at the indicated age
	- longitude  the geographic longitude of the location,
	             at the indicated age
	- value      the value to be set

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
var format string
var fromFlag float64
var toFlag float64
var atFlag float64

func setFlags(c *command.Command) {
	c.Flags().Float64Var(&fromFlag, "from", -1, "")
	c.Flags().Float64Var(&toFlag, "to", -1, "")
	c.Flags().Float64Var(&atFlag, "at", -1, "")
	c.Flags().StringVar(&inFlag, "in", "", "")
	c.Flags().StringVar(&format, "format", "timepix", "")
	c.Flags().StringVar(&format, "f", "timepix", "")
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

	tp, err := readTimePix(output, nil)
	if err != nil {
		return err
	}

	var stages []int64
	if atFlag >= 0 {
		stages = []int64{tp.ClosestStageAge(int64(atFlag * millionYears))}
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
	switch strings.ToLower(format) {
	case "latlon":
		ages := make(map[int64]bool, len(stages))
		for _, a := range stages {
			ages[a] = true
		}

		if err := addLocations(inFlag, tp, ages); err != nil {
			return err
		}
	case "timepix":
		source, err := readTimePix(inFlag, tp.Pixelation())
		if err != nil {
			return err
		}

		setTimeValue(tp, source, stages)
	default:
		return fmt.Errorf("unknown format %q", format)
	}

	if err := writeTimePix(output, tp); err != nil {
		return err
	}
	return nil
}

func setTimeValue(tp, source *model.TimePix, ages []int64) {
	for _, a := range ages {
		r := source.Stage(a)
		if r == nil {
			continue
		}
		for pix := 0; pix < tp.Pixelation().Len(); pix++ {
			v, ok := r[pix]
			if !ok {
				continue
			}
			if v == 0 {
				tp.Del(a, pix)
				continue
			}
			tp.Set(a, pix, v)
		}
	}
}

func readTimePix(name string, pix *earth.Pixelation) (*model.TimePix, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tp, err := model.ReadTimePix(f, pix)
	if err != nil {
		return nil, fmt.Errorf("when reading file %q: %v", name, err)
	}
	return tp, nil
}

var locHead = []string{
	"age",
	"latitude",
	"longitude",
	"value",
}

func addLocations(name string, tp *model.TimePix, ages map[int64]bool) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	tab := csv.NewReader(f)
	tab.Comma = '\t'
	tab.Comment = '#'

	head, err := tab.Read()
	if err != nil {
		return fmt.Errorf("file %q: header: %v", name, err)
	}
	fields := make(map[string]int, len(head))
	for i, h := range head {
		h = strings.ToLower(h)
		fields[h] = i
	}
	for _, h := range locHead {
		if _, ok := fields[h]; !ok {
			return fmt.Errorf("file %q: expecting field %q", name, h)
		}
	}

	pix := tp.Pixelation()
	for {
		row, err := tab.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		ln, _ := tab.FieldPos(0)
		if err != nil {
			return fmt.Errorf("on file %q: row %d: %v", name, ln, err)
		}

		f := "age"
		age, err := strconv.ParseInt(row[fields[f]], 10, 64)
		if err != nil {
			return fmt.Errorf("on file %q: row %d: field %q: %v", name, ln, f, err)
		}
		if !ages[age] {
			continue
		}

		f = "latitude"
		lat, err := strconv.ParseFloat(row[fields[f]], 64)
		if err != nil {
			return fmt.Errorf("on file %q: row %d: field %q: %v", name, ln, f, err)
		}
		if lat < -90 || lat > 90 {
			return fmt.Errorf("on file %q: row %d: field %q: invalid latitude value %.6f", name, ln, f, lat)
		}

		f = "longitude"
		lon, err := strconv.ParseFloat(row[fields[f]], 64)
		if err != nil {
			return fmt.Errorf("on file %q: row %d: field %q: %v", name, ln, f, err)
		}
		if lon < -180 || lon > 180 {
			return fmt.Errorf("on file %q: row %d: field %q: invalid longitude value %.6f", name, ln, f, lon)
		}

		f = "value"
		v, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return fmt.Errorf("on file %q: row %d: field %q: %v", name, ln, f, err)
		}

		px := pix.Pixel(lat, lon).ID()
		if v == 0 {
			tp.Del(age, px)
			continue
		}
		tp.Set(age, px, v)
	}
	return nil
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
