// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package ids implement a command to retrieve the pixel IDs
// of all pixels in an equal area pixelation.
package ids

import (
	"fmt"

	"github.com/js-arias/command"
	"github.com/js-arias/earth"
)

var Command = &command.Command{
	Usage: "ids [-e|--equator <value>]",
	Short: "print pixel IDs",
	Long: `
Command ids prints the IDs and locations of all pixels in a pixelation based
on an equal area partitioning of a sphere.

The IDs and locations will be printed in the standard output as tab-delimited
values, with the following columns:
	
	id         the ID of the pixel.
	latitude   the latitude of the pixel.
	longitude  the longitude of the pixel.

By default, the pixelation will be 360 pixels at the equator. Use the flag
--equator, or -e, to define a different pixelation.
	`,
	SetFlags: setFlags,
	Run:      run,
}

var equator int

func setFlags(c *command.Command) {
	c.Flags().IntVar(&equator, "equator", 360, "")
	c.Flags().IntVar(&equator, "e", 360, "")
}

func run(c *command.Command, args []string) error {
	pix := earth.NewPixelation(equator)

	fmt.Fprintf(c.Stdout(), "pixel\tlat\tlon\n")
	for id := 0; id < pix.Len(); id++ {
		pt := pix.ID(id).Point()
		fmt.Fprintf(c.Stdout(), "%d\t%.6f\t%.6f\n", id, pt.Latitude(), pt.Longitude())
	}

	return nil
}
