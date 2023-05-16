// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package pixel implements a command to get the a pixel location
// in a pixelation based on an equal area partitioning.
package pixel

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/js-arias/command"
	"github.com/js-arias/earth"
	"github.com/js-arias/earth/vector"
)

var Command = &command.Command{
	Usage: "pixel [-e|--equator <value>] [--id] [<value>...]",
	Short: "get pixel location",
	Long: `
Command pixel retrieves a pixel location in a pixelation based on an equal
area partitioning of a sphere for one or more values given as arguments.

Values to be retrieved will be read as arguments. If no argument is given then
values will be read from the standard input, one value per line, ignoring lines
starting with '#' character.

By default the values are reads as coordinates, with the first argument being
the latitude and the second the latitude. Note that they must be separated
arguments. If the first latitude is negative use "--" before the value
(otherwise the value will be interpreted as a flag). For coordinates the pixel
ID in the equal area pixelation will be retrieved. If the flag --id is
defined, then values will be read as pixel IDs in the pixelation, and the
geographic coordinates of the central point of the pixel will be retrieved.

By default the pixelation will be of 360 pixels at the equator. Use the flag
--equator, or -e, to define a different pixelation.
	`,
	SetFlags: setFlags,
	Run:      run,
}

var equator int
var idFlag bool

func setFlags(c *command.Command) {
	c.Flags().BoolVar(&idFlag, "id", false, "")
	c.Flags().IntVar(&equator, "equator", 360, "")
	c.Flags().IntVar(&equator, "e", 360, "")
}

func run(c *command.Command, args []string) error {
	pix := earth.NewPixelation(equator)
	if idFlag {
		var ids []int

		if len(args) == 0 {
			var err error
			ids, err = inPixels(c.Stdin(), pix.Len())
			if err != nil {
				return err
			}
		} else {
			for _, a := range args {
				id, err := readPixID(a, pix.Len())
				if err != nil {
					return err
				}
				ids = append(ids, id)
			}
		}

		fmt.Fprintf(c.Stdout(), "pixel\tlat\tlon\n")
		for _, id := range ids {
			pt := pix.ID(id).Point()
			fmt.Fprintf(c.Stdout(), "%d\t%.6f\t%.6f\n", id, pt.Latitude(), pt.Longitude())
		}
		return nil
	}

	var pts []vector.Point
	if len(args) == 0 {
		var err error
		pts, err = inLatLon(c.Stdin())
		if err != nil {
			return err
		}
	} else {
		if len(args)%2 != 0 {
			return fmt.Errorf("invalid number of coordinates: %d", len(args))
		}
		for i := 0; i < len(args); i += 2 {
			pt, err := vector.ParsePoint(args[i], args[i+1])
			if err != nil {
				return err
			}
			pts = append(pts, pt)
		}
	}

	fmt.Fprintf(c.Stdout(), "lat\tlon\tpixel\n")
	for _, pt := range pts {
		id := pix.Pixel(pt.Lat, pt.Lon).ID()
		fmt.Fprintf(c.Stdout(), "%.6f\t%.6f\t%d\n", pt.Lat, pt.Lon, id)
	}

	return nil
}

func inLatLon(in io.Reader) ([]vector.Point, error) {
	var pts []vector.Point

	r := bufio.NewReader(in)
	for i := 1; ; i++ {
		ln, err := r.ReadString('\n')
		if ln == "" && err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("at line %d: %v", i, err)
		}

		if ln == "" {
			continue
		}
		if ln[0] == '#' {
			continue
		}
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		v := strings.Fields(ln)
		if len(v) < 2 {
			return nil, fmt.Errorf("at line %d: invalid value %q: expecting \"lat lon\"", i, ln)
		}
		pt, err := vector.ParsePoint(v[0], v[1])
		if err != nil {
			return nil, fmt.Errorf("at line %d: %v", i, err)
		}
		pts = append(pts, pt)
	}
	return pts, nil
}

func inPixels(in io.Reader, max int) ([]int, error) {
	var ids []int

	r := bufio.NewReader(in)
	for i := 1; ; i++ {
		ln, err := r.ReadString('\n')
		if ln == "" && err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("at line %d: %v", i, err)
		}

		if ln == "" {
			continue
		}
		if ln[0] == '#' {
			continue
		}
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}

		id, err := readPixID(ln, max)
		if err != nil {
			return nil, fmt.Errorf("at line %d: %v", i, err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func readPixID(s string, max int) (int, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid value %q: %v", s, err)
	}
	if v >= max {
		return 0, fmt.Errorf("invalid value %q: invalid pixel", s)
	}
	return v, nil
}
