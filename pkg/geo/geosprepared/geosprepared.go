// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package geosprepared

import (
	"hash/fnv"

	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

// Cache is an instance of a GEOS Cache.
type Cache struct {
	mu struct {
		lock  syncutil.Mutex
		cache map[uint64]geos.PreparedGeometry
	}
}

// Free releases all resources in the cache.
func (c *Cache) Free() {
	c.mu.lock.Lock()
	defer c.mu.lock.Unlock()

	for _, g := range c.mu.cache {
		_ /* err */ = geos.PreparedGeomDestroy(g)
	}
	c.mu.cache = nil
}

func key(b geopb.EWKB) uint64 {
	h := fnv.New64()
	h.Write([]byte(b))
	return h.Sum64()
}

// Get will attempt to fetch from the cache the prepared geometry
// for either a or b, returning the prepared geometry of one side
// and the spatial object for the other.
// If neither is found, it will attempt to prepare one of either
// a or b.
// If neither shape should be cached, it will return a nil
// PreparedGeometry.
func (c *Cache) Get(
	a geopb.SpatialObject, b geopb.SpatialObject,
) (geos.PreparedGeometry, geopb.SpatialObject, error) {
	c.mu.lock.Lock()
	defer c.mu.lock.Unlock()

	if c.mu.cache == nil {
		c.mu.cache = make(map[uint64]geos.PreparedGeometry)
	}

	var aKey uint64
	var aValid bool

	switch a.ShapeType {
	case geopb.ShapeType_Point, geopb.ShapeType_MultiPoint:
	default:
		aKey = key(a.EWKB)
		if prep, ok := c.mu.cache[aKey]; ok {
			return prep, b, nil
		}
		aValid = true
	}

	var bKey uint64
	var bValid bool

	switch b.ShapeType {
	case geopb.ShapeType_Point, geopb.ShapeType_MultiPoint:
	default:
		bKey = key(b.EWKB)
		if prep, ok := c.mu.cache[bKey]; ok {
			return prep, a, nil
		}
		bValid = true
	}

	// Now determine which one to cache. If both are valid, choose the
	// longer EWKB.
	if aValid && bValid {
		if len(a.EWKB) > len(b.EWKB) {
			p, err := geos.PrepareGeometry(a.EWKB)
			if err != nil {
				return nil, geopb.SpatialObject{}, err
			}
			c.mu.cache[aKey] = p
			return p, b, nil
		}
		p, err := geos.PrepareGeometry(b.EWKB)
		if err != nil {
			return nil, geopb.SpatialObject{}, err
		}
		c.mu.cache[bKey] = p
		return p, a, nil
	}
	if aValid {
		p, err := geos.PrepareGeometry(a.EWKB)
		if err != nil {
			return nil, geopb.SpatialObject{}, err
		}
		c.mu.cache[aKey] = p
		return p, b, nil
	}
	if bValid {
		p, err := geos.PrepareGeometry(b.EWKB)
		if err != nil {
			return nil, geopb.SpatialObject{}, err
		}
		c.mu.cache[bKey] = p
		return p, a, nil
	}
	return nil, geopb.SpatialObject{}, nil
}
