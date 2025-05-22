// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package values implement a command to print
// the pixel values defined in a time pixelation model.
package values

import (
	"fmt"
	"os"
	"slices"

	"github.com/js-arias/command"
	"github.com/js-arias/earth/model"
	"github.com/js-arias/earth/pixkey"
)

var Command = &command.Command{
	Usage: "values [--key <key-file>] <time-pix-file>",
	Short: "print pixel values of a time pixelation model",
	Long: `
Command values read a time pixelation model and print the pixel values defined
for the model.

The argument of the command is the name of the file that contains the time
pixelation model.

With the flag --key a key-file can be used to define the labels of the pixels
values. A key file is a tab-delimited file with the following required
columns:

	key	the value used as identifier
	color	an RGB value separated by commas,
		for example "125,132,148".
	label	the label for the pixel value. 

Any other column will be ignored. Here is an example of a key file:

	key	color	gray	label
	0	54, 75, 154	255	deep ocean
	1	74, 123, 183	235	oceanic plateaus
	2	152, 202, 225	225	continental shelf
	3	254, 218, 139	195	lowlands
	4	246, 126, 75	185	highlands
	5	231, 231, 231	245	ice sheets
	`,
	SetFlags: setFlags,
	Run:      run,
}

var keyFlag string

func setFlags(c *command.Command) {
	c.Flags().StringVar(&keyFlag, "key", "", "")
}

func run(c *command.Command, args []string) error {
	if len(args) < 1 {
		return c.UsageError("expecting time pixelation model file")
	}

	pv, err := readValues(args[0])
	if err != nil {
		return err
	}
	labels, err := readKey()
	if err != nil {
		return err
	}

	for _, v := range pv {
		l := labels[v]
		fmt.Fprintf(c.Stdout(), "%d\t%s\n", v, l)
	}
	return nil
}

func readValues(name string) ([]int, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tp, err := model.ReadTimePix(f, nil)
	if err != nil {
		return nil, fmt.Errorf("when reading file %q: %v", name, err)
	}

	val := make(map[int]bool)
	for _, age := range tp.Stages() {
		s := tp.Stage(age)
		for _, v := range s {
			val[v] = true
		}
	}
	val[0] = true

	pv := make([]int, 0, len(val))
	for v := range val {
		pv = append(pv, v)
	}
	slices.Sort(pv)

	return pv, nil
}

func readKey() (map[int]string, error) {
	labels := make(map[int]string)
	if keyFlag == "" {
		return labels, nil
	}
	pk, err := pixkey.Read(keyFlag)
	if err != nil {
		return nil, err
	}

	for _, k := range pk.Keys() {
		labels[k] = pk.Label(k)
	}
	return labels, nil
}
