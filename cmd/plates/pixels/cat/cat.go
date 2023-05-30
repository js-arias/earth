// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package cat implements a command to merge several plate pixelation files
// into a single file.
package cat

import (
	"fmt"
	"io"
	"os"

	"github.com/js-arias/command"
	"github.com/js-arias/earth/model"
)

var Command = &command.Command{
	Usage: "cat [-o|--output <file>] [<pix-file>...]",
	Short: "merge plate pixelation files",
	Long: `
Cat merge one or more pixelated plates files into a single plate pixelation
file.

One or more pixelated plates files can be given as arguments. If no files are
given, the input will be read from the standard input.

The resulting pixelation will be written to the standard output. Use the
--output or -o flag to specify an output file.
	`,
	SetFlags: setFlags,
	Run:      run,
}

var output string

func setFlags(c *command.Command) {
	c.Flags().StringVar(&output, "output", "", "")
	c.Flags().StringVar(&output, "o", "", "")
}

func run(c *command.Command, args []string) error {
	if len(args) == 0 {
		args = append(args, "-")
	}

	var pp *model.PixPlate
	for _, a := range args {
		np, err := readPixPlate(c.Stdin(), a)
		if err != nil {
			return err
		}
		if pp == nil {
			pp = np
			continue
		}
		addPixels(pp, np)
	}

	if pp == nil {
		return nil
	}

	if err := write(c.Stdout(), output, pp); err != nil {
		return err
	}
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

func addPixels(pp, src *model.PixPlate) {
	pix := src.Pixelation()
	for _, plate := range src.Plates() {
		for _, id := range src.Pixels(plate) {
			px := src.Pixel(plate, id)
			pt := pix.ID(id).Point()
			pp.Add(plate, px.Name, pt.Latitude(), pt.Longitude(), px.Begin, px.End)
		}
	}
}

func write(w io.Writer, name string, pp *model.PixPlate) (err error) {
	if name != "" {
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
		w = f
	} else {
		name = "stdout"
	}

	if err := pp.TSV(w); err != nil {
		return fmt.Errorf("when writing on file %q: %v", name, err)
	}
	return nil
}
