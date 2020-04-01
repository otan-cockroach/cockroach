// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

#include <stdlib.h>
#include <dlfcn.h>

#include "geos_unix.h"

// Data Types adapted from `capi/geos_c.h.in` in GEOS.
typedef void* GEOSHandle;
typedef void* GEOSWKTReader;
typedef void* GEOSWKBWriter;

// Function declarations from `capi/geos_c.h.in` in GEOS.
typedef GEOSHandle (*GEOS_init_r)();
typedef void (*GEOS_finish_r)(GEOSHandle);

typedef void (*GEOSGeom_destroy_r)(GEOSHandle, GEOSGeometry);

typedef GEOSWKTReader (*GEOSWKTReader_create_r)(GEOSHandle);
typedef GEOSGeometry (*GEOSWKTReader_read_r)(GEOSHandle, GEOSWKTReader, char*);
typedef void (*GEOSWKTReader_destroy_r)(GEOSHandle, GEOSWKTReader);

typedef GEOSWKBWriter (*GEOSWKBWriter_create_r)(GEOSHandle);
typedef unsigned char* (*GEOSWKBWriter_write_r)(GEOSHandle, GEOSWKBWriter, GEOSGeometry, size_t*);
typedef void (*GEOSWKBWriter_destroy_r)(GEOSHandle, GEOSWKBWriter);

struct CockroachGEOSLib {
  void *dlHandle;

  GEOS_init_r GEOS_init_r;
  GEOS_finish_r GEOS_finish_r;

  GEOSGeom_destroy_r GEOSGeom_destroy_r;

  GEOSWKTReader_create_r GEOSWKTReader_create_r;
  GEOSWKTReader_destroy_r GEOSWKTReader_destroy_r;
  GEOSWKTReader_read_r GEOSWKTReader_read_r;

  GEOSWKBWriter_create_r GEOSWKBWriter_create_r;
  GEOSWKBWriter_destroy_r GEOSWKBWriter_destroy_r;
  GEOSWKBWriter_write_r GEOSWKBWriter_write_r;
};

CockroachGEOSLib *CockroachGEOSBootstrap(char *libLocation) {
  char *error;

  void *dlHandle = dlopen(libLocation, RTLD_LAZY);
  if (!dlHandle) {
    return NULL;
  }

  CockroachGEOSLib *lib = (CockroachGEOSLib*) malloc(sizeof(struct CockroachGEOSLib));
  lib->dlHandle = dlHandle;

  // TODO(otan): autogenerate all of this.
  lib->GEOS_init_r = (GEOS_init_r) dlsym(dlHandle, "GEOS_init_r");
  if ((error = dlerror()) != NULL) {
    free(lib);
    return NULL;
  }
  lib->GEOS_finish_r = (GEOS_finish_r) dlsym(dlHandle, "GEOS_finish_r");
  if ((error = dlerror()) != NULL) {
    free(lib);
    return NULL;
  }

  lib->GEOSGeom_destroy_r = (GEOSGeom_destroy_r) dlsym(dlHandle, "GEOSGeom_destroy_r");
  if ((error = dlerror()) != NULL) {
    free(lib);
    return NULL;
  }

  lib->GEOSWKTReader_create_r = (GEOSWKTReader_create_r) dlsym(dlHandle, "GEOSWKTReader_create_r");
  if ((error = dlerror()) != NULL) {
    free(lib);
    return NULL;
  }
  lib->GEOSWKTReader_destroy_r = (GEOSWKTReader_destroy_r) dlsym(dlHandle, "GEOSWKTReader_destroy_r");
  if ((error = dlerror()) != NULL) {
    free(lib);
    return NULL;
  }
  lib->GEOSWKTReader_read_r = (GEOSWKTReader_read_r) dlsym(dlHandle, "GEOSWKTReader_read_r");
  if ((error = dlerror()) != NULL) {
    free(lib);
    return NULL;
  }

  lib->GEOSWKBWriter_create_r = (GEOSWKBWriter_create_r) dlsym(dlHandle, "GEOSWKBWriter_create_r");
  if ((error = dlerror()) != NULL) {
    free(lib);
    return NULL;
  }
  lib->GEOSWKBWriter_destroy_r = (GEOSWKBWriter_destroy_r) dlsym(dlHandle, "GEOSWKBWriter_destroy_r");
  if ((error = dlerror()) != NULL) {
    free(lib);
    return NULL;
  }
  lib->GEOSWKBWriter_write_r = (GEOSWKBWriter_write_r) dlsym(dlHandle, "GEOSWKBWriter_write_r");
  if ((error = dlerror()) != NULL) {
    free(lib);
    return NULL;
  }

  return lib;
}

unsigned char* CockroachGEOSWKTToWKB(CockroachGEOSLib *lib, char *wkt, size_t *size) {
  unsigned char *result = NULL;

  GEOSHandle handle = lib->GEOS_init_r();
  GEOSWKTReader wktReader = lib->GEOSWKTReader_create_r(handle);
  GEOSGeometry geom = lib->GEOSWKTReader_read_r(handle, wktReader, wkt);
  lib->GEOSWKTReader_destroy_r(handle, wktReader);

  if (geom != NULL) {
    GEOSWKBWriter wkbWriter = lib->GEOSWKBWriter_create_r(handle);
    result = lib->GEOSWKBWriter_write_r(handle, wkbWriter, geom, size);
    lib->GEOSWKBWriter_destroy_r(handle, wkbWriter);
    lib->GEOSGeom_destroy_r(handle, geom);
  }

  lib->GEOS_finish_r(handle);
  return result;
}
