// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package lencmd implements a command to get the number of pixels
// in an isolatitude pixelation.
package lencmd

import (
	"fmt"

	"github.com/js-arias/command"
	"github.com/js-arias/earth"
)

var Command = &command.Command{
	Usage: "len [-e|--equator <value>] [--rings]",
	Short: "get the number of pixels in a pixelation",
	Long: `
Command len retrieves the number of pixels produced by an isolatitude
pixelation with the given number of pixels at the equator, as well as the
number of isolatitude rings.

By default the pixelation will be of 360 pixels at the equator. Use the flag
--equator, or -e, to define a different pixelation.

If the flag --rings is defined, the number of pixels at each ring will be
printed.
	`,
	SetFlags: setFlags,
	Run:      run,
}

var equator int
var rings bool

func setFlags(c *command.Command) {
	c.Flags().BoolVar(&rings, "rings", false, "")
	c.Flags().IntVar(&equator, "equator", 360, "")
	c.Flags().IntVar(&equator, "e", 360, "")
}

func run(c *command.Command, args []string) error {
	pix := earth.NewPixelation(equator)
	fmt.Fprintf(c.Stdout(), "pixels: %d\nrings: %d\n", pix.Len(), pix.Rings())
	if rings {
		for i := 0; i < pix.Rings(); i++ {
			fmt.Fprintf(c.Stdout(), "ring %d [lat %.6f]: %d\n", i, pix.RingLat(i), pix.PixPerRing(i))
		}
	}

	return nil
}
