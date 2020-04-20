// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package geogfn

import (
	"fmt"
	"math"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/golang/geo/s2"
)

var epsilon = 2e-16

// Covers returns whether geography A covers geography B.
// This calculation is done on the sphere.
func Covers(a *geo.Geography, b *geo.Geography) (bool, error) {
	aRegions, err := a.AsS2()
	if err != nil {
		return false, err
	}
	bRegions, err := b.AsS2()
	if err != nil {
		return false, err
	}
	// The S2 contains function almost does what we want - the problem is that
	// it does *not* consider the exterior of a geography to be part of the covering.
	// As such, we've had to define our own definition for most rather than
	// being able to directly rely on the S2 primitives.
	for _, aRegion := range aRegions {
		switch aRegion := aRegion.(type) {
		case s2.Point:
			for _, bRegion := range bRegions {
				switch bRegion := bRegion.(type) {
				case s2.Point:
					if !aRegion.ContainsPoint(bRegion) {
						return false, nil
					}
				case *s2.Polyline:
					return false, nil
				case *s2.Polygon:
					return false, nil
				default:
					return false, fmt.Errorf("unknown s2 type of b: %#v", bRegion)
				}
			}
		case *s2.Polyline:
			for _, bRegion := range bRegions {
				switch bRegion := bRegion.(type) {
				case s2.Point:
					return polylineCoversPoint(aRegion, bRegion), nil
				case *s2.Polyline:
					if !polylineCoversPolyline(aRegion, bRegion) {
						return false, nil
					}
				case *s2.Polygon:
					return false, nil
				default:
					return false, fmt.Errorf("unknown s2 type of b: %#v", bRegion)
				}
			}
		case *s2.Polygon:
			for _, bRegion := range bRegions {
				switch bRegion := bRegion.(type) {
				case s2.Point:
					return polygonCoversPoint(aRegion, bRegion), nil
				case *s2.Polyline:
					if !polygonCoversPolyline(aRegion, bRegion) {
						return false, nil
					}
				case *s2.Polygon:
					if !polygonCoversPolygon(aRegion, bRegion) {
						return false, nil
					}
				default:
					return false, fmt.Errorf("unknown s2 type of b: %#v", bRegion)
				}
			}
		default:
			return false, fmt.Errorf("unknown s2 type of a: %#v", aRegion)
		}
	}
	return true, nil
}

// CoveredBy returns whether geography A is covered by geography B.
// This calculation is done on the sphere.
func CoveredBy(a *geo.Geography, b *geo.Geography) (bool, error) {
	return Covers(b, a)
}

// polylineCoversPoints returns whether a polyline covers a given point.
func polylineCoversPoint(a *s2.Polyline, b s2.Point) bool {
	covers, _ := polylineCoversPointWithIdx(a, b)
	return covers
}

// polylineCoversPointsWithIdx returns whether a polyline covers a given point.
// If true, it will also return an index of which point there was an intersection.
func polylineCoversPointWithIdx(a *s2.Polyline, b s2.Point) (bool, int) {
	for edgeIdx := 0; edgeIdx < a.NumEdges(); edgeIdx++ {
		if edgeCoversPoint(a.Edge(edgeIdx), b) {
			return true, edgeIdx
		}
	}
	return false, -1
}

// polygonCoversPoints returns whether a polygon covers a given point.
func polygonCoversPoint(a *s2.Polygon, b s2.Point) bool {
	return a.ContainsPoint(b)
}

// edgeCoversPoint determines whether a given edge contains a point.
// We "cheat" here, and use the same precision as an s2 cell, which is a cm.
// This should be good enough for geography based calculations.
func edgeCoversPoint(e s2.Edge, p s2.Point) bool {
	a := s2.PointFromLatLng(s2.LatLngFromDegrees(1, 1))
	b := s2.PointFromLatLng(s2.LatLngFromDegrees(2, 2))
	c := s2.PointFromLatLng(s2.LatLngFromDegrees(2, 1))
	d := s2.PointFromLatLng(s2.LatLngFromDegrees(1, 2))
	intersect := s2.Intersection(a, b, c, d)
	fmt.Printf("%s\n", s2.LatLngFromPoint(intersect))
	if e.V0 == p || e.V1 == p {
		return true
	}
	return edgeConeHasPoint(e, p) && edgePlaneHasPoint(e, p)
}

// edgePlaneHasPoint returns true if the point is on the plane defined by the edge.
func edgePlaneHasPoint(e s2.Edge, p s2.Point) bool {
	normal := e.V0.Cross(e.V1.Vector).Normalize()
	return math.Abs(normal.Dot(p.Vector)) < epsilon
}

// edgeConeHasPoint returns true if the point is within a "cone" defined by the edge.
// To do this, calculate the center of the edge and normalize it.
// If the projection of the point onto the center is larger than the start, then it is
// inside the given cone.
func edgeConeHasPoint(e s2.Edge, p s2.Point) bool {
	center := e.V0.Add(e.V1.Vector).Normalize()
	projectStartToCenter := e.V0.Dot(center)
	projectPointToCenter := p.Dot(center)
	return projectPointToCenter > projectStartToCenter || projectStartToCenter-projectPointToCenter < epsilon
}

// polylineCoversPolyline returns whether polyline a covers polyline b.
func polylineCoversPolyline(a *s2.Polyline, b *s2.Polyline) bool {
	aCoversStartOfB, aCoverBStart := polylineCoversPointWithIdx(a, (*b)[0])
	// We must first check that the start and end of B is contained by A.
	if !aCoversStartOfB || !polylineCoversPoint(a, (*b)[len(*b)-1]) {
		return false
	}
	// The algorithm is as follows:
	// * Find when B starts to be a "subline" of "A".
	// * While B is part of line "A", make sure each point is contained
	//   in a line of A.
	// * Finish when no more points in line A or B.
	aPoints := *a
	bPoints := *b
	aIdx := aCoverBStart
	bIdx := 0

	for aIdx < len(aPoints)-1 && bIdx < len(bPoints)-1 {
		// moved indicates whether we've made forward progress in the lookup.
		moved := false

		aEdge := s2.Edge{V0: aPoints[aIdx], V1: aPoints[aIdx+1]}
		bEdge := s2.Edge{V0: bPoints[bIdx], V1: bPoints[bIdx+1]}

		// If the edge of A contains the start of edge B, move B forward.
		if edgeCoversPoint(aEdge, bEdge.V0) {
			bIdx++
			moved = true
		}

		// If the edge of B contains the start of edge A, move A forward.
		if edgeCoversPoint(bEdge, aEdge.V0) {
			aIdx++
			moved = true
		}

		// Since we have not been able to make progress, we fail since something along
		// the way did not match.
		if !moved {
			return false
		}
	}

	// At this point, everything in B is found to be within A. Since
	// we checked that the start and end of B
	return true
}

// polygonCoversPolyline returns whether polygon a covers polyline b.
func polygonCoversPolyline(a *s2.Polygon, b *s2.Polyline) bool {
	// Check everything of polyline B is within polygon A.
	for _, point := range *b {
		if !polygonCoversPoint(a, point) {
			return false
		}
	}
	// Even if every point of polyline B is inside polygon A, they
	// may form an edge which goes outside of polygon A and back in
	// due to geodetic lines.
	//
	// As such, check if there are any intersections - if so,
	// we do not consider it a covering.
	return !polygonIntersectsPolyline(a, b)
}

// polygonIntersectsPolyline returns whether polygon a intersects with
// polygon b.
// We do not consider edges which are touching / the same as intersecting.
func polygonIntersectsPolyline(a *s2.Polygon, b *s2.Polyline) bool {
	// Avoid using NumEdges / Edge of the Polygon type as it is not O(1).
	for _, loop := range a.Loops() {
		for loopEdgeIdx := 0; loopEdgeIdx < loop.NumEdges(); loopEdgeIdx++ {
			loopEdge := loop.Edge(loopEdgeIdx)
			crosser := s2.NewChainEdgeCrosser(loopEdge.V0, loopEdge.V1, (*b)[0])
			for _, nextVertex := range (*b)[1:] {
				if crosser.ChainCrossingSign(nextVertex) == s2.Cross {
					return true
				}
			}
		}
	}
	return false
}

// polygonCoversPolygon returns whether polygon a intersects with polygon b.
func polygonCoversPolygon(a *s2.Polygon, b *s2.Polygon) bool {
	// We can rely on Contains here, as if the exteriors are the same it will
	// still return true in here.
	return a.Contains(b)
}
