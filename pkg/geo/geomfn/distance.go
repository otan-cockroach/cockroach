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
	"github.com/cockroachdb/errors"
	"github.com/twpayne/go-geom"
)

// minDistanceInternal finds the minimum distance between two geometries.
// This implementation is done in-house, as compared to using GEOS.
func minDistanceInternal(a *geo.Geometry, b *geo.Geometry, stopAfterLE float64) (float64, error) {
	aGeoms, err := flattenGeometry(a)
	if err != nil {
		return 0, err
	}
	bGeoms, err := flattenGeometry(b)
	if err != nil {
		return 0, err
	}
	f := newMinDistanceFinder(stopAfterLE)
	if _, err := distanceInternal(aGeoms, bGeoms, f); err != nil {
		return 0, err
	}
	return f.distance(), nil
}

// distanceInternal returns the distance from two Geometries based on the distanceValueUpdater.
// Assumes geometries have been flattened.
func distanceInternal(
	aGeoms []geom.T, bGeoms []geom.T, distanceValueUpdater distanceValueUpdater,
) (bool, error) {
	// TODO(otan): optimize by using bboxes to avoid certain checks.
	for _, a := range aGeoms {
		for _, b := range bGeoms {
			switch a := a.(type) {
			case *geom.Point:
				switch b := b.(type) {
				case *geom.Point:
					if distanceValueUpdater.updateDistanceValue(pointFromGeomPoint(a), pointFromGeomPoint(b)) {
						return true, nil
					}
				case *geom.LineString:
					if distanceInternalGeomPointToGeomLine(a, b, distanceValueUpdater) {
						return true, nil
					}
				case *geom.Polygon:
					if distanceInternalGeomPointToGeomPolygon(a, b, distanceValueUpdater) {
						return true, nil
					}
				default:
					return false, errors.Newf("unknown geom type: %T", b)
				}
			case *geom.LineString:
				switch b := b.(type) {
				case *geom.Point:
					if distanceInternalGeomPointToGeomLine(b, a, distanceValueUpdater) {
						return true, nil
					}
				case *geom.LineString:
					if distanceInternalGeomCoordsToGeomCoords(a, b, distanceValueUpdater) {
						return true, nil
					}
				case *geom.Polygon:
				default:
					return false, errors.Newf("unknown geom type: %T", b)
				}
			case *geom.Polygon:
				switch b := b.(type) {
				case *geom.Point:
					if distanceInternalGeomPointToGeomPolygon(b, a, distanceValueUpdater) {
						return true, nil
					}
				case *geom.LineString:
				case *geom.Polygon:
				default:
					return false, errors.Newf("unknown geom type: %T", b)
				}
			default:
				return false, errors.Newf("unknown geom type: %T", a)
			}
		}
	}
	return false, nil
}

// distanceInternalGeomPointToGeomLine calculates the distance from a point to a line, returning
// if the calling function should exit early.
func distanceInternalGeomPointToGeomLine(
	a *geom.Point, b *geom.LineString, distanceValueUpdater distanceValueUpdater,
) bool {
	p := pointFromGeomPoint(a)
	return distanceInternalPointToGeomCoords(p, b, distanceValueUpdater)
}

// geomCoords is an interface around any object in go-geom that has a coord attribute.
type geomCoords interface {
	Coord(i int) geom.Coord
	NumCoords() int
}

// Edge returns the i-th edge of the geomCoords.
func edgeFromGeomCoords(g geomCoords, i int) edge {
	return edge{V0: pointFromGeomCoord(g.Coord(i)), V1: pointFromGeomCoord(g.Coord(i + 1))}
}

var _ geomCoords = (*geom.LineString)(nil)
var _ geomCoords = (*geom.LinearRing)(nil)

// distanceInternalPointToGeomCoords returns the distance between a point and a geom object
// that has a number of coords.
func distanceInternalPointToGeomCoords(
	p point, b geomCoords, distanceValueUpdater distanceValueUpdater,
) bool {
	if distanceValueUpdater.updateDistanceValue(p, pointFromGeomCoord(b.Coord(0))) {
		return true
	}
	for i := 0; i < b.NumCoords()-1; i++ {
		e := edgeFromGeomCoords(b, i)
		if distanceValueUpdater.updateDistanceValue(p, e.V1) {
			return true
		}
		if closestPoint, ok := e.maybeFindClosestPoint(p); ok {
			if distanceValueUpdater.updateDistanceValue(p, closestPoint) {
				return true
			}
		}
	}
	return false
}

// geomPolygonCoversPoint returns whether a polygon has the given point
// on the boundary or internally within a LinearRing.
// This is done using the winding number algorithm, also known as the
// "non-zero rule".
// See: https://en.wikipedia.org/wiki/Point_in_polygon
// See also: https://en.wikipedia.org/wiki/Winding_number
// See also: https://en.wikipedia.org/wiki/Nonzero-rule
func geomLinearRingCoversPoint(a *geom.LinearRing, p point) bool {
	windingNumber := 0
	for i := 0; i < a.NumCoords()-1; i++ {
		e := edgeFromGeomCoords(a, i)
		yMin := math.Min(e.V0.Y, e.V1.Y)
		yMax := math.Max(e.V0.Y, e.V1.Y)
		// If the edge isn't on the same level as Y, this edge isn't worth considering.
		if p.Y > yMax || p.Y < yMin {
			continue
		}
		side := findPointSide(p, e)
		// If the point is on the line, there is a covering.
		if side == pointSideOn && p.inBoundingCross(e) {
			return true
		}
		// If the point is left of the segment and the line is rising
		// we have a circle going CCW, so increment.
		if side == pointSideLeft && e.V0.Vector.Y <= p.Vector.Y && p.Vector.Y < e.V1.Vector.Y {
			windingNumber++
		}
		// If the line is to the right of the segment and the
		// line is falling, we a have a circle going CW so decrement.
		if side == pointSideRight && e.V1.Vector.Y <= p.Vector.Y && p.Vector.Y < e.V0.Vector.Y {
			windingNumber--
		}
	}
	return windingNumber != 0
}

// distanceInternalGeomPointToGeomPolygon calculates the distance from a point to a line, returning
// if the calling function should exit early.
func distanceInternalGeomPointToGeomPolygon(
	a *geom.Point, b *geom.Polygon, distanceValueUpdater distanceValueUpdater,
) bool {
	p := pointFromGeomPoint(a)
	// If the exterior ring does not cover the outer ring, we just need to calculate the distance
	// to the outer ring.
	if !geomLinearRingCoversPoint(b.LinearRing(0), p) {
		return distanceInternalPointToGeomCoords(p, b.LinearRing(0), distanceValueUpdater)
	}

	// At this point it may be inside a hole.
	// If it is in a hole, return the distance to the hole.
	for ringIdx := 1; ringIdx < b.NumLinearRings(); ringIdx++ {
		ring := b.LinearRing(ringIdx)
		if geomLinearRingCoversPoint(ring, p) {
			return distanceInternalPointToGeomCoords(p, ring, distanceValueUpdater)
		}
	}

	// Otherwise, we are inside the polygon.
	return distanceValueUpdater.updateDistanceValue(p, p)
}

// distanceInternalGeomCoord calculate the distance from one set of coords to another, returning
// if the calling function should exit early.
// It assumes these coordinates do not intersect.
func distanceInternalGeomCoordsToGeomCoords(
	a geomCoords, b geomCoords, distanceValueUpdater distanceValueUpdater,
) bool {
	for aIdx := 0; aIdx < a.NumCoords()-1; aIdx++ {
		aEdge := edgeFromGeomCoords(a, aIdx)
		for bIdx := 0; bIdx < b.NumCoords()-1; bIdx++ {
			bEdge := edgeFromGeomCoords(b, bIdx)
			// TODO(otan): copy edge crossing code from S2, and re-use it.
			// If two edges intersect, then their distance is zero.

			for _, c := range []struct {
				p point
				e edge
			}{
				{aEdge.V0, bEdge},
				{aEdge.V1, bEdge},
				{bEdge.V0, aEdge},
				{bEdge.V1, aEdge},
			} {
				if distanceValueUpdater.updateDistanceValue(c.p, c.e.V0) {
					return true
				}
				if distanceValueUpdater.updateDistanceValue(c.p, c.e.V1) {
					return true
				}

				if closestPoint, ok := c.e.maybeFindClosestPoint(c.p); ok {
					if distanceValueUpdater.updateDistanceValue(c.p, closestPoint) {
						return true
					}
				}
			}
		}
	}
	return false
}

// distanceValueUpdater includes hooks on distanceInternal that allow for calculating
// different required distances.
type distanceValueUpdater interface {
	updateDistanceValue(a point, b point) bool
	distance() float64
}

var _ distanceValueUpdater = (*minDistanceFinder)(nil)

// minDistanceFinder implements distanceValueUpdater to find the minimum distance.
type minDistanceFinder struct {
	currentValue float64
	stopAfterLE  float64
}

// newMinDistanceFinder returns a new evaluator that finds the minimum distance.
func newMinDistanceFinder(stopAfterLE float64) *minDistanceFinder {
	return &minDistanceFinder{currentValue: math.MaxFloat64, stopAfterLE: stopAfterLE}
}

// minDistance implements distanceValueUpdater..
func (f *minDistanceFinder) distance() float64 {
	return math.Sqrt(f.currentValue)
}

// updateDistanceValue implements distanceValueUpdater.
func (f *minDistanceFinder) updateDistanceValue(a point, b point) bool {
	distance := b.Vector.Sub(a.Vector).Norm()
	if distance <= f.currentValue {
		f.currentValue = distance
		return f.currentValue <= f.stopAfterLE
	}
	return false
}
