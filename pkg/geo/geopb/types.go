// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package geopb

// SRID is a Spatial Reference Identifer. All geometry and geography figures are
// stored and represented as bare floats. SRIDs tie these floats to the planar
// or spherical coordinate system, allowing them to be intrepred and compared.
//
// The zero value is special and means an unknown coordinate system.
type SRID int32

// WKT is the Well Known Text form of a spatial object.
type WKT string

// WKB is the Well Known Bytes form of a spatial object.
type WKB []byte

// EWKB is the Extended Well Known Bytes form of a spatial object.
type EWKB []byte
