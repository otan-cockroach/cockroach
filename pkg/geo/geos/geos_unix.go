// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// +build !windows

package geos

import (
	"os"
	"path"
	"runtime"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/errors"
)

// #cgo LDFLAGS: -ldl
// #include "geos_unix.h"
import "C"

// geosHolder is the struct that contains all dlsym loaded functions.
type geosHolder struct {
	lib *C.CockroachGEOSLib
}

// validOrError returns an error if the geosHolder is not valid.
func (gh *geosHolder) validOrError() error {
	if gh == nil {
		// TODO(otan): be more helpful in this error message.
		return errors.Newf("could not load GEOS library")
	}
	return nil
}

// geos is the global instance of geosHolder, which is initialized at init time.
var holder = geosHolder{}

// defaultGEOSLocations contains a list of locations where GEOS is expected to exist.
// TODO(otan): make this configurable by flags.
var defaultGEOSLocations = []string{
	// TODO: put mac / linux locations
}

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if runtime.GOOS == "Darwin" {
		defaultGEOSLocations = append(defaultGEOSLocations, path.Join(cwd, "../../../lib/libgeos_c.dylib"))
	} else {
		defaultGEOSLocations = append(defaultGEOSLocations, path.Join(cwd, "../../../lib/libgeos_c.so"))
	}
	holder = bootstrap(defaultGEOSLocations)
}

// bootstrap initializes a geosHolder by attempting to dlopen all the
// paths as parsed in by locs.
func bootstrap(locs []string) geosHolder {
	var h geosHolder
	for _, loc := range locs {
		cLoc := C.CString(loc)
		lib := C.CockroachGEOSBootstrap(cLoc)
		if lib != nil {
			holder.lib = lib
			break
		}
	}
	return h
}

// WKTToWKB parses a WKT into WKB using the GEOS library.
func WKTToWKB(wkt geo.WKT) (geo.WKB, error) {
	if err := holder.validOrError(); err != nil {
		return nil, err
	}
	var dataLen C.size_t
	cWKT := C.CString(string(wkt))
	defer C.free(unsafe.Pointer(cWKT))
	cWKB := C.CockroachGEOSWKTToWKB(holder.lib, cWKT, &dataLen)
	if cWKB == nil {
		return nil, errors.Newf("error decoding WKT: %s", wkt)
	}
	defer C.free(unsafe.Pointer(cWKB))

	wkb := geo.WKB(C.GoBytes(unsafe.Pointer(cWKB), C.int(dataLen)))
	return wkb, nil
}
