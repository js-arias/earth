// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package list implements a command to list plates
// from a plate pixelation file.
package list

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"

	"github.com/js-arias/command"
	"github.com/js-arias/earth/model"
)

var Command = &command.Command{
	Usage: "list [--len] [<pix-file>...]",
	Short: "list plates from a file with pixelated plates",
	Long: `
List reads one or more pixelated plates files and prints a list of the plates,
including the names of the features and their time intervals.

One or more input files can be given as arguments. If no files are given, the
input will be read from the standard input.
	
If the --len flag is specified, it will print the number of pixels for each
feature.
	`,
	SetFlags: setFlags,
	Run:      run,
}

var lenFlag bool

func setFlags(c *command.Command) {
	c.Flags().BoolVar(&lenFlag, "len", false, "")
}

func run(c *command.Command, args []string) error {
	if len(args) == 0 {
		args = append(args, "-")
	}

	pd := make(map[int]*plateData)
	for _, a := range args {
		pp, err := readPixPlate(c.Stdin(), a)
		if err != nil {
			return err
		}
		addFeatures(pp, pd)
	}

	printFeatures(c.Stdout(), pd)
	return nil
}

func readPixPlate(r io.Reader, name string) (*model.PixPlate, error) {
	if name != "-" {
		f, err := os.Open(name)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r = f
	} else {
		name = "stdin"
	}

	pp, err := model.ReadPixPlate(r, nil)
	if err != nil {
		return nil, fmt.Errorf("when reading file %q: %v", name, err)
	}
	return pp, nil
}

func addFeatures(pp *model.PixPlate, pd map[int]*plateData) {
	for _, plate := range pp.Plates() {
		p, ok := pd[plate]
		if !ok {
			p = &plateData{
				plate:    plate,
				features: make(map[string]*feature),
			}
			pd[plate] = p
		}

		pix := pp.Pixels(plate)
		for _, id := range pix {
			px := pp.Pixel(plate, id)
			name := px.Name
			if name == "" {
				name = strconv.Itoa(plate)
			}
			f, ok := p.features[name]
			if !ok {
				f = &feature{
					name:  name,
					begin: px.Begin,
					end:   px.End,
					size:  1,
				}
				p.features[name] = f
				continue
			}

			if px.Begin > f.begin {
				f.begin = px.Begin
			}
			if px.End < f.end {
				f.end = px.End
			}
			f.size++
		}
	}
}

type plateData struct {
	plate    int
	features map[string]*feature
}

type feature struct {
	name  string
	begin int64
	end   int64
	size  int
}

// MillionYears is used to transform ages
// from an integer number of years
// to a float in million years.
const millionYears = 1_000_000

func printFeatures(w io.Writer, pd map[int]*plateData) {
	plates := make([]int, 0, len(pd))
	for _, p := range pd {
		plates = append(plates, p.plate)
	}
	slices.Sort(plates)

	for _, plate := range plates {
		p := pd[plate]
		names := make([]string, 0, len(p.features))
		for _, f := range p.features {
			names = append(names, f.name)
		}
		slices.Sort(names)

		for _, nm := range names {
			f := p.features[nm]
			fmt.Fprintf(w, "%d\t%s\t%.6f\t%.6f", plate, nm, float64(f.begin)/millionYears, float64(f.end)/millionYears)
			if lenFlag {
				fmt.Fprintf(w, "\t%d", f.size)
			}
			fmt.Fprintf(w, "\n")
		}
	}
}
