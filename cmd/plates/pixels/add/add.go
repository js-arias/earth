// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package add implements a command to add locations
// to a plate pixelation.
package add

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/js-arias/command"
	"github.com/js-arias/earth/model"
)

var Command = &command.Command{
	Usage: "add <pix-file> [<location-file>...]",
	Short: "add locations to a plate pixelation file",
	Long: `
Add reads a file with pixelated plates and add one or more files with
locations and plates to the pixelation.

The plate pixelation file must be given as the first argument of the command.
	
One or more location files can be given as arguments. If no files are given,
the input will be read from the standard input.
	
The locations file is a tab-delimited text file with the following columns:
	
	- plate      the plate ID for the added location
	- latitude   the geographic latitude of the location
	- longitude  the geographic longitude of the location
	- name       the name of the tectonic feature; this field is optional
	- begin      the oldest age of the feature in years
	- end        the youngest age of the feature in years
	`,
	Run: run,
}

func run(c *command.Command, args []string) error {
	if len(args) == 0 {
		return c.UsageError("expecting <pix-file>")
	}
	ppFile := args[0]
	pp, err := readPixPlate(ppFile)
	if err != nil {
		return err
	}

	args = args[1:]
	if len(args) == 0 {
		args = append(args, "-")
	}
	for _, a := range args {
		if err := addLocations(c.Stdin(), a, pp); err != nil {
			return err
		}
	}

	if err := write(ppFile, pp); err != nil {
		return err
	}
	return nil
}

func readPixPlate(name string) (*model.PixPlate, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pp, err := model.ReadPixPlate(f, nil)
	if err != nil {
		return nil, fmt.Errorf("when reading file %q: %v", name, err)
	}
	return pp, nil
}

var locHead = []string{
	"plate",
	"latitude",
	"longitude",
	"begin",
	"end",
}

func addLocations(r io.Reader, name string, pp *model.PixPlate) error {
	if name != "-" {
		f, err := os.Open(name)
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	} else {
		name = "stdin"
	}

	tab := csv.NewReader(r)
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

	for {
		row, err := tab.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		ln, _ := tab.FieldPos(0)
		if err != nil {
			return fmt.Errorf("on file %q: row %d: %v", name, ln, err)
		}

		f := "plate"
		plate, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return fmt.Errorf("on file %q: row %d: field %q: %v", name, ln, f, err)
		}

		f = "latitude"
		lat, err := strconv.ParseFloat(row[fields[f]], 64)
		if err != nil {
			return fmt.Errorf("on file %q: row %d: field %q: %v", name, ln, f, err)
		}

		f = "longitude"
		lon, err := strconv.ParseFloat(row[fields[f]], 64)
		if err != nil {
			return fmt.Errorf("on file %q: row %d: field %q: %v", name, ln, f, err)
		}

		f = "begin"
		begin, err := strconv.ParseInt(row[fields[f]], 10, 64)
		if err != nil {
			return fmt.Errorf("on file %q; row %d: field %q: %v", name, ln, f, err)
		}
		f = "end"
		end, err := strconv.ParseInt(row[fields[f]], 10, 64)
		if err != nil {
			return fmt.Errorf("on file %q: row %d: field %q: %v", name, ln, f, err)
		}
		if end > begin {
			return fmt.Errorf("on file %q: row %d: field %q: end value must be less than %d", name, ln, f, begin)
		}

		plateName := ""
		f = "name"
		if c, ok := fields[f]; ok {
			plateName = row[c]
		}

		pp.Add(plate, plateName, lat, lon, begin, end)
	}
	return nil
}

func write(name string, pp *model.PixPlate) (err error) {
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

	if err := pp.TSV(f); err != nil {
		return fmt.Errorf("when writing on file %q: %v", name, err)
	}
	return nil
}
