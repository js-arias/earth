// Copyright © 2022 J. Salvador Arias <jsalarias@gmail.com>
// All rights reserved.
// Distributed under BSD2 license that can be found in the LICENSE file.

// Package model implements paleogeographic reconstruction models
// using an isolatitude pixelation
// and discrete time stages.
//
// There are different model types depending
// on how is expected the reconstruction model will be used.
//
// To build a model,
// type PixPlate is used to provide associations between tectonic plates
// (which rotations are defined in a rotation file)
// and the pixels of that plate.
package model
