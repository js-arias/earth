// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// DistMat benchmarks the calculation of the spherical normal
// using ring distances and a distance matrix.
package main

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/js-arias/earth"
	"github.com/js-arias/earth/stat/dist"
)

type noMatChan struct {
	pixel  int
	pix    *earth.Pixelation
	pixels []int
	pdf    dist.Normal
	wg     *sync.WaitGroup
}

type answerChan struct {
	pixel int
	prob  float64
}

func noMatrixProb(nmC chan noMatChan, answer chan answerChan) {
	for c := range nmC {
		pt1 := c.pix.ID(c.pixel).Point()
		max := -math.MaxFloat64
		for _, px2 := range c.pixels {
			pt2 := c.pix.ID(px2).Point()
			dist := earth.Distance(pt1, pt2)
			p := c.pdf.LogProb(dist)
			if p > max {
				max = p
			}
		}

		answer <- answerChan{
			pixel: c.pixel,
			prob:  max,
		}
		c.wg.Done()
	}
}

var numCPU = runtime.NumCPU()

func noMatrix(pixels []int) {
	pix := earth.NewPixelation(360)
	n := dist.NewNormal(100, pix)

	send := make(chan noMatChan, numCPU*2)
	answer := make(chan answerChan, numCPU*2)
	for i := 0; i < numCPU; i++ {
		go noMatrixProb(send, answer)
	}

	start := time.Now()
	// send the pixels
	go func() {
		var wg sync.WaitGroup
		for _, px := range pixels {
			wg.Add(1)
			send <- noMatChan{
				pixel:  px,
				pix:    pix,
				pixels: pixels,
				pdf:    n,
				wg:     &wg,
			}
		}
		wg.Wait()
		close(answer)
	}()

	max := -math.MaxFloat64
	for a := range answer {
		if a.prob > max {
			max = a.prob
		}
	}
	close(send)

	fmt.Printf("no matrix: total time %v [max.val = %.6f]\n", time.Since(start), max)
}

type withMatChan struct {
	pixel  int
	dm     *earth.DistMat
	pixels []int
	pdf    dist.Normal
	wg     *sync.WaitGroup
}

func withMatrixProb(wmC chan withMatChan, answer chan answerChan) {
	for c := range wmC {
		max := -math.MaxFloat64
		for _, px2 := range c.pixels {
			p := c.pdf.LogProbRingDist(c.dm.At(c.pixel, px2))
			if p > max {
				max = p
			}
		}

		answer <- answerChan{
			pixel: c.pixel,
			prob:  max,
		}
		c.wg.Done()
	}
}

func withMatrix(pixels []int) {
	pix := earth.NewPixelation(360)

	start := time.Now()
	dm, err := earth.NewDistMat(pix)
	if err != nil {
		panic(err)
	}
	fmt.Printf("matrix build time: %v\n", time.Since(start))
	n := dist.NewNormal(100, pix)

	send := make(chan withMatChan, numCPU*2)
	answer := make(chan answerChan, numCPU*2)
	for i := 0; i < numCPU; i++ {
		go withMatrixProb(send, answer)
	}

	start = time.Now()
	// send the pixels
	go func() {
		var wg sync.WaitGroup
		for _, px := range pixels {
			wg.Add(1)
			send <- withMatChan{
				pixel:  px,
				dm:     dm,
				pixels: pixels,
				pdf:    n,
				wg:     &wg,
			}
		}
		wg.Wait()
		close(answer)
	}()

	max := -math.MaxFloat64
	for a := range answer {
		if a.prob > max {
			max = a.prob
		}
	}
	close(send)

	fmt.Printf("distance matrix: total time %v [max.val = %.6f]\n", time.Since(start), max)
}

func main() {
	pix := earth.NewPixelation(360)

	// make random pixels
	pixels := make([]int, 20000)
	for i := range pixels {
		pixels[i] = pix.Random().ID()
	}

	noMatrix(pixels)
	withMatrix(pixels)
}
