// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package importcmd implements a command to import features
// from a GPlates GPML file
// into an equal area pixelation.
package importcmd

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"

	"github.com/js-arias/command"
	"github.com/js-arias/earth"
	"github.com/js-arias/earth/model"
	"github.com/js-arias/earth/vector"
)

var Command = &command.Command{
	Usage: `import [-e|--equator <value>] [--at <age>]
	[--cpu <value>] [-o|--output <file>] [<gpml-file>...]`,
	Short: "import GPML files",
	Long: `
Import reads one or more GPML encoded GPlates files and imports them
into an equal area pixelation.

The GPML format is an implementation of the XML format for tectonic plate
modeling, primarily used by GPlates software (<https://www.gplates.org>). For
a formal description of the GPML format, refer to:
<https://www.gplates.org/docs/gpgim/>.

One or more input files can be given as arguments. If no files are specified,
the input will be read from the standard input.

By default, the new pixelation will have 360 pixels at the equator (i.e., a
one-degree pixelation). To change the number of pixels, use the --equator
or -e flag.

By default, all features will be pixelated. Use the --at flag to import only
features that existed at the specified time (in million years).

The resulting pixelation will be written to the standard output. Use the
--output or -o flag to specify an output file.

The output file is a tab-delimited value file with the following columns:

	- equator: the number of pixels at the equator
	- plate:   the ID of a tectonic plate
	- pixel:   the ID of a pixel (from an isolatitude pixelation)
	- name:    the name of a tectonic feature
	- begin:   the oldest age of the pixel (in years)
	- end:     the youngest age of the pixel (in years)

By default, the import process will utilize all available CPU processors
concurrently. Use the --cpu flag to set the number of used processors.
	`,
	SetFlags: setFlags,
	Run:      run,
}

var output string
var atFlag float64
var equator int
var cpu int

func setFlags(c *command.Command) {
	c.Flags().StringVar(&output, "output", "", "")
	c.Flags().StringVar(&output, "o", "", "")
	c.Flags().IntVar(&equator, "equator", 360, "")
	c.Flags().IntVar(&equator, "e", 360, "")
	c.Flags().IntVar(&cpu, "cpu", runtime.NumCPU(), "")
	c.Flags().Float64Var(&atFlag, "at", 0, "")
}

// MillionYears is used to transform age
// (a float in million years)
// to an integer in years.
const millionYears = 1_000_000

func run(c *command.Command, args []string) (err error) {
	features := make(chan vector.Feature)
	errChan := make(chan error)

	go read(c.Stdin(), args, features, errChan)

	pp := model.NewPixPlate(earth.NewPixelation(equator))

	done := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < cpu; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range features {
				pix := f.Pixels(pp.Pixelation())
				pp.AddPixels(f.Plate, f.Name, pix, f.Begin, f.End)
			}
		}()
	}
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err = <-errChan:
		return err
	case <-done:
	}

	if err := write(c.Stdout(), output, pp); err != nil {
		return err
	}
	return nil
}

func read(r io.Reader, args []string, fc chan vector.Feature, ec chan error) {
	at := int64(millionYears * atFlag)

	if len(args) == 0 {
		args = append(args, "-")
	}

	var wg sync.WaitGroup
	for _, a := range args {
		wg.Add(1)
		go func(a string) {
			defer wg.Done()

			fs, err := readFeatures(r, a)
			if err != nil {
				ec <- err
				return
			}
			for _, f := range fs {
				if at != 0 && (f.Begin < at || f.End > at) {
					continue
				}
				fc <- f
			}
		}(a)
	}

	wg.Wait()
	close(fc)
}

func readFeatures(r io.Reader, name string) ([]vector.Feature, error) {
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

	fs, err := vector.DecodeGPML(r)
	if err != nil {
		return nil, fmt.Errorf("while reading from %q: %v", name, err)
	}

	return fs, nil
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
