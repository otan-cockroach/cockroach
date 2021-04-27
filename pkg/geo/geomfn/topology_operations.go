// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package geomfn

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/cockroachdb/errors"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
)

// Boundary returns the boundary of a given Geometry.
func Boundary(g geo.Geometry) (geo.Geometry, error) {
	// follow PostGIS behavior
	if g.Empty() {
		return g, nil
	}
	boundaryEWKB, err := geos.Boundary(g.EWKB())
	if err != nil {
		return geo.Geometry{}, err
	}
	return geo.ParseGeometryFromEWKB(boundaryEWKB)
}

// Centroid returns the Centroid of a given Geometry.
func Centroid(g geo.Geometry) (geo.Geometry, error) {
	centroidEWKB, err := geos.Centroid(g.EWKB())
	if err != nil {
		return geo.Geometry{}, err
	}
	return geo.ParseGeometryFromEWKB(centroidEWKB)
}

// MinimumBoundingCircle returns minimum bounding circle of an EWKB
func MinimumBoundingCircle(g geo.Geometry) (geo.Geometry, geo.Geometry, float64, error) {
	polygonEWKB, centroidEWKB, radius, err := geos.MinimumBoundingCircle(g.EWKB())
	if err != nil {
		return geo.Geometry{}, geo.Geometry{}, 0, err
	}

	polygon, err := geo.ParseGeometryFromEWKB(polygonEWKB)
	if err != nil {
		return geo.Geometry{}, geo.Geometry{}, 0, err
	}

	centroid, err := geo.ParseGeometryFromEWKB(centroidEWKB)
	if err != nil {
		return geo.Geometry{}, geo.Geometry{}, 0, err
	}

	return polygon, centroid, radius, nil
}

// ClipByRect clips a given Geometry by the given BoundingBox.
func ClipByRect(g geo.Geometry, b geo.CartesianBoundingBox) (geo.Geometry, error) {
	if g.Empty() {
		return g, nil
	}
	clipByRectEWKB, err := geos.ClipByRect(g.EWKB(), b.LoX, b.LoY, b.HiX, b.HiY)
	if err != nil {
		return geo.Geometry{}, err
	}
	return geo.ParseGeometryFromEWKB(clipByRectEWKB)
}

// ConvexHull returns the convex hull of a given Geometry.
func ConvexHull(g geo.Geometry) (geo.Geometry, error) {
	convexHullEWKB, err := geos.ConvexHull(g.EWKB())
	if err != nil {
		return geo.Geometry{}, err
	}
	return geo.ParseGeometryFromEWKB(convexHullEWKB)
}

// Difference returns the difference between two given Geometries.
func Difference(a, b geo.Geometry) (geo.Geometry, error) {
	// follow PostGIS behavior
	if a.Empty() || b.Empty() {
		return a, nil
	}
	if a.SRID() != b.SRID() {
		return geo.Geometry{}, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	}
	diffEWKB, err := geos.Difference(a.EWKB(), b.EWKB())
	if err != nil {
		return geo.Geometry{}, err
	}
	return geo.ParseGeometryFromEWKB(diffEWKB)
}

// Simplify returns a simplified Geometry.
func Simplify(g geo.Geometry, tolerance float64, preserveCollapsed bool) (geo.Geometry, error) {
	if math.IsNaN(tolerance) || g.Empty() || g.ShapeType2D() == geopb.ShapeType_Point || g.ShapeType2D() == geopb.ShapeType_MultiPoint {
		return g, nil
	}
	simplifiedEWKB, err := geos.Simplify(g.EWKB(), tolerance)
	if err != nil {
		return geo.Geometry{}, err
	}
	t, err := ewkb.Unmarshal(simplifiedEWKB)
	if err != nil {
		return geo.Geometry{}, err
	}
	if t, err = processCollapsedAfterSimplify(t, preserveCollapsed); err != nil {
		return geo.Geometry{}, err
	}
	return geo.MakeGeometryFromGeomT(t)
}

func processCollapsedAfterSimplify(
	original geo.Geometry, t geom.T, preserveCollapsed bool,
) (geom.T, error) {
	switch t := t.(type) {
	case *geom.Point, *geom.MultiPoint:
		return t, nil
	case *geom.LineString:
		if simplifyShouldCollapseLineString(t, preserveCollapsed) {
			return geom.NewLineString(t.Layout()).SetSRID(t.SRID()), nil
		}
		return t, nil
	case *geom.MultiLineString:
		needsChange := false
		// First scan to avoid the extra allocations if unnecessary.
		for i := 0; i < t.NumLineStrings(); i++ {
			if simplifyShouldCollapseLineString(t.LineString(i)) {
				needsChange = true
				break
			}
		}
		if !needsChange {
			return t, nil
		}
		ret := geom.NewMultiLineString(t.Layout()).SetSRID(t.SRID())
		for i := 0; i < t.NumLineStrings(); i++ {
			toAppend, err := processCollapsedAfterSimplify(t, preserveCollapsed)
			if err != nil {
				return nil, err
			}
			if err := ret.Push(toAppend); err != nil {
				return nil, err
			}
			return ret, nil
		}
	case *geom.Polygon:
		if simplifyShouldUncollapsePolygon(t, preserveCollapsed) {

		}
		return t, nil
	case *geom.GeometryCollection:
		ret := geom.NewGeometryCollection().SetSRID(t.SRID())
		for _, gcT := range t.Geoms() {
			pushT, err := processCollapsedAfterSimplify(gcT, preserveCollapsed)
			if err != nil {
				return nil, err
			}
			if err := ret.Push(pushT); err != nil {
				return nil, err
			}
		}
		return ret, nil
	default:
		return nil, errors.Newf("unknown geom.T after simplify: %T", t)
	}
}

// simplifyShouldCollapseLineString determines whether a LineString should be
// collapsed.
// GEOS will make two equal points if the LineString is fully collapsed.
// We should collapse the LineString if it is requested.
func simplifyShouldCollapseLineString(t *geom.LineString, preserveCollapsed bool) bool {
	return t.NumPoints() == 2 && !preserveCollapsed && t.Coord(0).Equal(geom.XY, t.Coord(1))
}

// simplifyShouldUncollapsePolygon determines whether a Polygon should be
// uncollapsed.
// GEOS will return POLYGON EMPTY if a point is collapsed.
func simplifyShouldUncollapsePolygon(t *geom.Polygon, preserveCollapsed bool) bool {
	return t.IsEmpty() && preserveCollapsed
}

// SimplifyPreserveTopology returns a simplified Geometry with topology preserved.
func SimplifyPreserveTopology(g geo.Geometry, tolerance float64) (geo.Geometry, error) {
	simplifiedEWKB, err := geos.TopologyPreserveSimplify(g.EWKB(), tolerance)
	if err != nil {
		return geo.Geometry{}, err
	}
	return geo.ParseGeometryFromEWKB(simplifiedEWKB)
}

// PointOnSurface returns the PointOnSurface of a given Geometry.
func PointOnSurface(g geo.Geometry) (geo.Geometry, error) {
	pointOnSurfaceEWKB, err := geos.PointOnSurface(g.EWKB())
	if err != nil {
		return geo.Geometry{}, err
	}
	return geo.ParseGeometryFromEWKB(pointOnSurfaceEWKB)
}

// Intersection returns the geometries of intersection between A and B.
func Intersection(a geo.Geometry, b geo.Geometry) (geo.Geometry, error) {
	if a.SRID() != b.SRID() {
		return geo.Geometry{}, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	}
	// Match PostGIS.
	if a.Empty() {
		return a, nil
	}
	if b.Empty() {
		return b, nil
	}
	retEWKB, err := geos.Intersection(a.EWKB(), b.EWKB())
	if err != nil {
		return geo.Geometry{}, err
	}
	return geo.ParseGeometryFromEWKB(retEWKB)
}

// UnaryUnion returns the geometry of union between input geometry components
func UnaryUnion(g geo.Geometry) (geo.Geometry, error) {
	retEWKB, err := geos.UnaryUnion(g.EWKB())
	if err != nil {
		return geo.Geometry{}, err
	}
	return geo.ParseGeometryFromEWKB(retEWKB)
}

// Union returns the geometries of union between A and B.
func Union(a geo.Geometry, b geo.Geometry) (geo.Geometry, error) {
	if a.SRID() != b.SRID() {
		return geo.Geometry{}, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	}
	retEWKB, err := geos.Union(a.EWKB(), b.EWKB())
	if err != nil {
		return geo.Geometry{}, err
	}
	return geo.ParseGeometryFromEWKB(retEWKB)
}

// SymDifference returns the geometries of symmetric difference between A and B.
func SymDifference(a geo.Geometry, b geo.Geometry) (geo.Geometry, error) {
	if a.SRID() != b.SRID() {
		return geo.Geometry{}, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	}
	retEWKB, err := geos.SymDifference(a.EWKB(), b.EWKB())
	if err != nil {
		return geo.Geometry{}, err
	}
	return geo.ParseGeometryFromEWKB(retEWKB)
}

// SharedPaths Returns a geometry collection containing paths shared by the two input geometries.
func SharedPaths(a geo.Geometry, b geo.Geometry) (geo.Geometry, error) {
	if a.SRID() != b.SRID() {
		return geo.Geometry{}, geo.NewMismatchingSRIDsError(a.SpatialObject(), b.SpatialObject())
	}
	paths, err := geos.SharedPaths(a.EWKB(), b.EWKB())
	if err != nil {
		return geo.Geometry{}, err
	}
	gm, err := geo.ParseGeometryFromEWKB(paths)
	if err != nil {
		return geo.Geometry{}, err
	}
	return gm, nil
}

// MinimumRotatedRectangle Returns a minimum rotated rectangle enclosing a geometry
func MinimumRotatedRectangle(g geo.Geometry) (geo.Geometry, error) {
	paths, err := geos.MinimumRotatedRectangle(g.EWKB())
	if err != nil {
		return geo.Geometry{}, err
	}
	gm, err := geo.ParseGeometryFromEWKB(paths)
	if err != nil {
		return geo.Geometry{}, err
	}
	return gm, nil
}
