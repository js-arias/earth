// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package variance implements a command to report
// the variance of a spherical normal.
package variance

import (
	"fmt"
	"strconv"

	"github.com/js-arias/command"
	"github.com/js-arias/earth"
	"github.com/js-arias/earth/stat/dist"
)

var Command = &command.Command{
	Usage: "variance [-e|--equator <value>] <lambda-value>",
	Short: "report variance of spherical normal",
	Long: `
Command variance prints the variance, in radians^2, for a given lambda value
of a spherical normal.

The spherical normal is defined by the lambda parameter (the concentration),
which is analogous to the kappa parameter (of the von Mises-Fisher
distribution) or the precision (of a planar normal). While in the planar
normal, the variance is just the inverse of the precision, that is not the
case for the spherical normal.
	
The argument of the command is the lambda value.
	
By default, the calculation is done using a pixelation with 360 pixels at the
equator. Use the flag --equator, or -e, to change the size of the pixelation.
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
	if len(args) == 0 {
		c.UsageError("expecting lambda value")
	}

	lambda, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid lambda value: %v", err)
	}

	pix := earth.NewPixelation(equator)
	n := dist.NewNormal(lambda, pix)
	fmt.Fprintf(c.Stdout(), "%.6f\n", n.Variance())

	return nil
}
