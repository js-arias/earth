// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

package vector

import (
	"image"
	"image/color"
	"math"

	"github.com/js-arias/earth"
	"golang.org/x/exp/slices"
	rasterizer "golang.org/x/image/vector"
)

// Pixels return an slice
// with the OD of pixels in a pixelation
// that are part of a feature.
func (f Feature) Pixels(pix *earth.Pixelation) []int {
	r := &raster{
		pix:    pix,
		pixels: make(map[int]bool),
	}

	if f.Point != nil {
		px := pix.Pixel(f.Point.Lat, f.Point.Lon).ID()
		r.pixels[px] = true
	}

	r.doRaster(f.Polygon)
	return r.pixSet()
}

type raster struct {
	pix    *earth.Pixelation
	pixels map[int]bool
}

func (r *raster) pixSet() []int {
	pix := make([]int, 0, len(r.pixels))
	for px := range r.pixels {
		pix = append(pix, px)
	}
	slices.Sort(pix)
	return pix
}

func (r *raster) doRaster(poly Polygon) {
	cols := r.pix.Equator() * 10

	north, south := poly.bounds()
	img := &azimuthal{
		hemisphere: hemisphere(north, south),
		cols:       cols,
		pixels:     make([]bool, cols*cols),
		radius:     float64(cols) / (2 * math.Pi),
		center:     float64(cols) / 2,
	}

	ras := rasterizer.NewRasterizer(cols, cols)
	for i, p := range poly {
		x, y := img.xy(p.Lat, p.Lon)
		if i == 0 {
			ras.MoveTo(float32(x), float32(y))
			continue
		}
		ras.LineTo(float32(x), float32(y))
	}

	src := &filled{cols}
	ras.Draw(img, img.Bounds(), src, image.Pt(0, 0))

	for pos, v := range img.pixels {
		if !v {
			continue
		}
		lat, lon := img.latLon(pos)
		px := r.pix.Pixel(lat, lon).ID()
		r.pixels[px] = true
		if lat > north {
			north = lat
		}
		if lat < south {
			south = lat
		}
	}

	for px := 0; px < r.pix.Len(); px++ {
		pt := r.pix.ID(px).Point()
		if pt.Latitude() > north {
			continue
		}
		if pt.Latitude() < south {
			continue
		}

		x, y := img.xy(pt.Latitude(), pt.Longitude())
		pos := int(x)*img.cols + int(y)
		if img.pixels[pos] {
			r.pixels[px] = true
		}
	}
}

// Hemisphere returns true for the northern hemisphere
// and false for the southern hemisphere.
func hemisphere(north, south float64) bool {
	if south == -90 {
		return false
	}
	if north == 90 {
		return true
	}

	if north < 0 {
		return false
	}
	if south > 0 {
		return true
	}

	return north < math.Abs(south)
}

type azimuthal struct {
	hemisphere bool
	cols       int
	pixels     []bool

	radius float64
	center float64
}

func (a *azimuthal) ColorModel() color.Model { return color.RGBAModel }
func (a *azimuthal) Bounds() image.Rectangle { return image.Rect(0, 0, a.cols, a.cols) }
func (a *azimuthal) At(x, y int) color.Color {
	pos := x*a.cols + y
	if a.pixels[pos] {
		return color.RGBA{0, 0, 0, 255}
	}
	return color.RGBA64{255, 255, 255, 255}
}

func (a *azimuthal) Set(x, y int, c color.Color) {
	pos := x*a.cols + y
	r, g, b, alpha := c.RGBA()
	if r > 100 || g > 100 || b > 100 || alpha < 100 {
		a.pixels[pos] = false
		return
	}
	a.pixels[pos] = true
}

func (a *azimuthal) xy(lat, lon float64) (x, y float64) {
	nLat := 90 - lat
	if !a.hemisphere {
		nLat = lat + 90
	}

	rho := a.radius * earth.ToRad(nLat)
	theta := earth.ToRad(lon)

	x = rho * math.Sin(theta)
	y = -rho * math.Cos(theta)

	return x + a.center, y + a.center
}

func (a *azimuthal) latLon(pos int) (lat, lon float64) {
	x := float64(pos/a.cols) + 0.5 - a.center
	y := float64(pos%a.cols) + 0.5 - a.center

	rho := math.Hypot(x, y)
	nLat := earth.ToDegree(rho / a.radius)
	lat = 90 - nLat
	if !a.hemisphere {
		lat = nLat - 90
	}
	theta := math.Asin(x / rho)
	lon = earth.ToDegree(theta)
	if y > 0 {
		if lon > 0 {
			lon = 180 - lon
		} else {
			lon = -180 - lon
		}
	}
	return lat, lon
}

type filled struct {
	cols int
}

func (f filled) ColorModel() color.Model { return color.RGBAModel }
func (f filled) Bounds() image.Rectangle { return image.Rect(0, 0, f.cols, f.cols) }
func (f filled) At(x, y int) color.Color { return color.RGBA{0, 0, 0, 255} }
