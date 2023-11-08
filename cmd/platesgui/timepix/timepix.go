// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package timepix implements a command to view and edit
// a time pixelation model.
package timepix

import (
	"encoding/csv"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/js-arias/blind"
	"github.com/js-arias/command"
	"github.com/js-arias/earth/model"
)

var Command = &command.Command{
	Usage: "timepix [--key <key-file>] <time-pix-file>",
	Short: "view and edit a time pixelation model",
	Long: `
Command timepix reads a time pixelation model from a file and displays it using
a plate carrée projection.

The argument of the command is the name of the file that contains the time
pixelation model.
	
In the display, all pixels with a given value will have the same color
(selected at random). With the flag --key, a key file can be used to define
the color used in the display. A key file is a tab-delimited file with the
following required columns:
	
	key    the value used as an identifier.
	color  an RBA value separated by commas, for example "125,132,148".
	
All other columns will be ignored. Here is an example of a key file:
	
	key	color	gray	comment
	0	54, 75, 154	255	deep ocean
	1	74, 123, 183	235	oceanic plateaus
	2	152, 202, 225	225	continental shelf
	3	254, 218, 139	195	lowlands
	4	246, 126, 75	185	highlands
	5	231, 231, 231	245	ice sheets

In this case, the gray and comment columns will be ignored.

At the bottom of the display, a status bar will show information about the
model. In the first field, a star "[*]" will be displayed if the model has
been modified. The second field is the time for the current stage, in million
years. The third field is the geographic location. The fourth field is the
pixel identifier. The fifth field is the value of the current pixel. The last
field, Set, shows the value used to set a pixel.

The following keys can be used:

	"N"  go to next time stage
	"P"  go to previous time stage
	"+"  zoom in
	"-"  zoom out
	"S"  changes the set value for a pixel
	"M"  shows a mask for all the pixels with the same value as 
	     the current pixel
	"W"  writes any change to the time pixelation model

To set a pixel, click the mouse over a pixel while holding the <shift> key.

Use a mouse drag to change the location of the map in the display.
	`,
	SetFlags: setFlags,
	Run:      run,
}

var keyFlag string

func setFlags(c *command.Command) {
	c.Flags().StringVar(&keyFlag, "key", "", "")
}

// MillionYears is used to transform ages
// (a float in million years)
// to an integer in years.
const millionYears = 1_000_000

type mapStagePix struct {
	pt     f32.Point   // current point
	offset f32.Point   // offset for the map origin
	box    image.Point // space for the map
	cols   int         // size of the map
	step   float64     // scale of the map
	mask   bool
	dirty  bool
	name   string // file name

	mVal int   // value used for the mask
	kv   int   // index of the value to set
	kvs  []int // values
	keys map[int]color.RGBA

	lat, lon float64
	stage    int // index of the current stage
	stages   []int64
	tp       *model.TimePix
}

func run(c *command.Command, args []string) error {
	if len(args) == 0 {
		return c.UsageError("expecting time pixelation model")
	}
	output := args[0]

	tp, err := readTimePix(args[0])
	if err != nil {
		return err
	}

	var keys map[int]color.RGBA
	if keyFlag != "" {
		keys, err = readKey(keyFlag)
		if err != nil {
			return err
		}
	} else {
		keys = makeKeyPalette(tp)
	}

	sp := mapStagePix{
		step: 0.5,
		cols: 720,
		name: output,

		kvs:  keyValues(keys),
		keys: keys,

		lat:    math.NaN(),
		lon:    math.NaN(),
		stages: tp.Stages(),
		tp:     tp,
	}

	go func() {
		w := app.NewWindow(
			app.Title("Time Pixelation Viewer-Editor"),
			app.Size(unit.Dp(720), unit.Dp(380)),
		)
		if sp.run(w); err != nil {
			fmt.Fprintf(c.Stderr(), "%s: %v.\n", "timepix", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
	return nil
}

func (sp *mapStagePix) run(w *app.Window) error {
	th := material.NewTheme()

	var ops op.Ops
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			sp.box = image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)
			events(gtx, sp)
			draw(gtx, th, sp)
			registerEvents(gtx, sp)
			e.Frame(gtx.Ops)
		case system.DestroyEvent:
			return e.Err
		}
	}
}

func draw(gtx layout.Context, th *material.Theme, sp *mapStagePix) {
	layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceStart,
	}.Layout(gtx,
		layout.Rigid(
			func(gtx layout.Context) layout.Dimensions {
				age := float64(sp.stages[sp.stage]) / millionYears
				pixID := "--"
				val := "--"
				if !math.IsNaN(sp.lat) {
					pix := sp.tp.Pixelation().Pixel(sp.lat, sp.lon).ID()
					pixID = strconv.Itoa(pix)
					v, _ := sp.tp.At(sp.stages[sp.stage], pix)
					val = strconv.Itoa(v)
				}
				dirty := ""
				if sp.dirty {
					dirty = "*"
				}
				coord := fmt.Sprintf("[%s] time: %.3f Ma, lat: %.2f lon: %.2f, pix: %s, val: %s, set to: %d", dirty, age, sp.lat, sp.lon, pixID, val, sp.kvs[sp.kv])
				status := material.Label(th, 12, coord)
				status.Alignment = text.Start

				d := status.Layout(gtx)
				sp.box.Y -= d.Size.Y
				return d
			},
		),
	)

	paint.NewImageOp(sp).Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}

func registerEvents(gtx layout.Context, sp *mapStagePix) {
	area := clip.Rect(image.Rect(0, 0, sp.box.X, sp.box.Y)).Push(gtx.Ops)

	pointer.InputOp{
		Tag:   sp,
		Types: pointer.Move | pointer.Drag | pointer.Press,
	}.Add(gtx.Ops)

	key.InputOp{
		Tag:  sp,
		Keys: key.Set("+|-|M|N|P|S|W"),
	}.Add(gtx.Ops)

	area.Pop()
}

func events(gtx layout.Context, sp *mapStagePix) {
	for _, e := range gtx.Events(sp) {
		switch e := e.(type) {
		case key.Event:
			if e.State != key.Press {
				continue
			}
			switch e.Name {
			case "+":
				if sp.step > 0.11 {
					sp.step -= 0.05
				} else {
					sp.step -= 0.01
				}
				if sp.step < 0.01 {
					sp.step = 0.01
				}
				sp.cols = int(360 / sp.step)
				if sp.cols%2 != 0 {
					sp.cols++
				}
				sp.setLocation()
			case "-":
				if sp.step < 0.1 {
					sp.step += 0.01
				} else {
					sp.step += 0.05
				}
				sp.cols = int(360 / sp.step)
				if sp.cols%2 != 0 {
					sp.cols++
				}
				sp.setLocation()
			case "M":
				if math.IsNaN(sp.lat) {
					continue
				}
				pix := sp.tp.Pixelation().Pixel(sp.lat, sp.lon).ID()
				sp.mVal, _ = sp.tp.At(sp.stages[sp.stage], pix)
				sp.mask = !sp.mask
			case "S":
				sp.kv++
				if sp.kv >= len(sp.kvs) {
					sp.kv = 0
				}
			case "W":
				if !sp.dirty {
					continue
				}
				if err := writeTimePix(sp.name, sp.tp); err != nil {
					fmt.Fprintf(os.Stderr, "Unable to write file: %v", err)
					continue
				}
				sp.dirty = false
			case "N":
				sp.stage++
				if sp.stage >= len(sp.stages) {
					sp.stage = len(sp.stages) - 1
				}
			case "P":
				sp.stage--
				if sp.stage < 0 {
					sp.stage = 0
				}
			}
		case pointer.Event:
			switch e.Type {
			case pointer.Drag:
				sp.offset.X += e.Position.X - sp.pt.X
				sp.offset.Y += e.Position.Y - sp.pt.Y
				sp.pt = e.Position
			case pointer.Move:
				sp.pt = e.Position
				sp.setLocation()
			case pointer.Press:
				if e.Modifiers&key.ModShift == 0 {
					continue
				}
				sp.setLocation()
				if math.IsNaN(sp.lat) {
					continue
				}
				pix := sp.tp.Pixelation().Pixel(sp.lat, sp.lon).ID()
				sp.tp.Set(sp.stages[sp.stage], pix, sp.kvs[sp.kv])
				sp.dirty = true
			}
		}
	}
}

func (sp *mapStagePix) setLocation() {
	y := sp.pt.Y - sp.offset.Y
	lat := 90 - float64(y)*sp.step
	if lat < -90 || lat >= 90 {
		sp.lat = math.NaN()
		sp.lon = math.NaN()
		return
	}
	x := sp.pt.X - sp.offset.X
	lon := float64(x)*sp.step - 180
	if lon < -180 || lon > 180 {
		sp.lat = math.NaN()
		sp.lon = math.NaN()
		return
	}
	sp.lat = lat
	sp.lon = lon

}

func (sp mapStagePix) ColorModel() color.Model { return color.RGBAModel }
func (sp mapStagePix) Bounds() image.Rectangle {
	cols := int(sp.offset.X) + sp.cols
	if cols > sp.box.X {
		cols = sp.box.X
	}
	rows := int(sp.offset.Y) + sp.cols/2
	if rows > sp.box.Y {
		rows = sp.box.Y
	}
	return image.Rect(0, 0, cols, rows)
}
func (sp mapStagePix) At(x, y int) color.Color {
	x = x - int(sp.offset.X)
	y = y - int(sp.offset.Y)
	if x < 0 || y < 0 {
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}

	lat := 90 - float64(y)*sp.step
	lon := float64(x)*sp.step - 180

	pix := sp.tp.Pixelation().Pixel(lat, lon).ID()
	v, _ := sp.tp.At(sp.stages[sp.stage], pix)
	if sp.mask {
		if sp.mVal == v {
			return color.RGBA{R: 255, G: 255, B: 255, A: 255}
		}
		return color.RGBA{A: 255}
	}
	c, ok := sp.keys[v]
	if !ok {
		return color.RGBA{A: 255}
	}
	return c
}

func readTimePix(name string) (*model.TimePix, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tp, err := model.ReadTimePix(f, nil)
	if err != nil {
		return nil, fmt.Errorf("when reading file %q: %v", name, err)
	}
	return tp, nil
}

func readKey(name string) (map[int]color.RGBA, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = '\t'
	r.Comment = '#'

	head, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("while reading file %q: while reading header: %v", name, err)
	}
	fields := make(map[string]int, len(head))
	for i, h := range head {
		h = strings.ToLower(h)
		fields[h] = i
	}
	for _, h := range []string{"key", "color"} {
		if _, ok := fields[h]; !ok {
			return nil, fmt.Errorf("while reading file %q: expecting field %q", name, h)
		}
	}

	keys := make(map[int]color.RGBA)
	for {
		row, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		ln, _ := r.FieldPos(0)
		if err != nil {
			return nil, fmt.Errorf("while reading file %q: on row %d: %v", name, ln, err)
		}

		f := "key"
		k, err := strconv.Atoi(row[fields[f]])
		if err != nil {
			return nil, fmt.Errorf("while reading file %q: on row %d: %v", name, ln, err)
		}

		f = "color"
		vals := strings.Split(row[fields[f]], ",")
		if len(vals) != 3 {
			return nil, fmt.Errorf("while reading file %q: on row %d: field %q: found %d values", name, ln, f, len(vals))
		}

		red, err := strconv.Atoi(strings.TrimSpace(vals[0]))
		if err != nil {
			return nil, fmt.Errorf("while reading file %q: on row %d: field %q [red value]: %v", name, ln, f, err)
		}
		if red > 255 {
			return nil, fmt.Errorf("while reading file %q: on row %d: field %q [red value]: invalid value %d", name, ln, f, red)
		}

		green, err := strconv.Atoi(strings.TrimSpace(vals[1]))
		if err != nil {
			return nil, fmt.Errorf("while reading file %q: on row %d: field %q [green value]: %v", name, ln, f, err)
		}
		if green > 255 {
			return nil, fmt.Errorf("while reading file %q: on row %d: field %q [green value]: invalid value %d", name, ln, f, green)
		}

		blue, err := strconv.Atoi(strings.TrimSpace(vals[2]))
		if err != nil {
			return nil, fmt.Errorf("while reading file %q: on row %d: field %q [blue value]: %v", name, ln, f, err)
		}
		if blue > 255 {
			return nil, fmt.Errorf("while reading file %q: on row %d: field %q [blue value]: invalid value %d", name, ln, f, blue)
		}

		c := color.RGBA{uint8(red), uint8(green), uint8(blue), 255}
		keys[k] = c
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("while reading file %q: %v", name, io.EOF)
	}
	return keys, nil
}

func makeKeyPalette(tp *model.TimePix) map[int]color.RGBA {
	keys := make(map[int]color.RGBA)
	for _, a := range tp.Stages() {
		for px := 0; px < tp.Pixelation().Len(); px++ {
			v, _ := tp.At(a, px)
			if _, ok := keys[v]; ok {
				continue
			}
			keys[v] = randColor()
		}
	}
	return keys
}

func randColor() color.RGBA {
	return blind.Sequential(blind.Iridescent, rand.Float64())
}

func keyValues(keys map[int]color.RGBA) []int {
	kvs := make([]int, 0, len(keys))
	for k := range keys {
		kvs = append(kvs, k)
	}
	slices.Sort(kvs)

	return kvs
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
