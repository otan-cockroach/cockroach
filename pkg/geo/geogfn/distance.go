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
	"github.com/cockroachdb/cockroach/pkg/geo/geographiclib"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

type useSphereOrSpheroid bool

const wgs84SphereRadius = 6371008.7714150598325213222

const (
	UseSpheroid useSphereOrSpheroid = true
	UseSphere   useSphereOrSpheroid = false
)

// Distance returns the distance between geographies a and b on a sphere or spheroid.
func Distance(
	a *geo.Geography, b *geo.Geography, useSpheroid useSphereOrSpheroid,
) (float64, error) {
	aRegions, err := a.AsS2()
	if err != nil {
		return 0, err
	}
	bRegions, err := b.AsS2()
	if err != nil {
		return 0, err
	}

	if useSpheroid {
		return minDistanceSpheroid(aRegions, bRegions, 0)
	}
	return minDistanceSphere(aRegions, bRegions)
}

// s2RegionsToPointsAndShapeIndexes separates S2 regions into either being part of the ShapeIndex,
// or an array of points depending on their type.
func s2RegionsToPointsAndShapeIndexes(regions []s2.Region) (*s2.ShapeIndex, []s2.Point, error) {
	shapeIndex := s2.NewShapeIndex()
	points := []s2.Point{}
	for _, region := range regions {
		switch region := region.(type) {
		case s2.Point:
			points = append(points, region)
		case *s2.Polygon:
			shapeIndex.Add(region)
		case *s2.Polyline:
			shapeIndex.Add(region)
		default:
			return nil, nil, fmt.Errorf("invalid region: %#v", region)
		}
	}
	return shapeIndex, points, nil
}

// minChordAngle returns the minimum chord angle of a and b.
func minChordAngle(a s1.ChordAngle, b s1.ChordAngle) s1.ChordAngle {
	if a < b {
		return a
	}
	return b
}

// spheroidInverse returns the s12 (meter) component of spheroid.Inverse from s2 Points.
func spheroidInverse(s geographiclib.Spheroid, a s2.Point, b s2.Point) float64 {
	inv, _, _ := s.Inverse(s2.LatLngFromPoint(a), s2.LatLngFromPoint(b))
	return inv
}

// maybeClosestPointToEdge returns a point on the normal of the edge,
// and an indication of whether it lies in the segment of the edge.
// The point on the normal may be outside the segment, hence the "maybe".
func maybeClosestPointToEdge(edge s2.Edge, point s2.Point) (s2.Point, bool) {
	edgeNormal := edge.V0.Vector.Cross(edge.V1.Vector).Normalize()
	edgeNormalScaledToPoint := edgeNormal.Mul(edgeNormal.Dot(point.Vector))
	closestPoint := s2.Point{Vector: point.Vector.Sub(edgeNormalScaledToPoint).Normalize()}
	return closestPoint, (&s2.Polyline{edge.V0, edge.V1}).IntersectsCell(s2.CellFromPoint(closestPoint))
}

// minSpheroidDistanceToEdgeEndAndProjection calculates the distance between a point
// and the end of the Edge, as well as a projection of the point on the Edge.
func minSpheroidDistancePointToEdgeEndAndProjection(
	minDistance float64,
	s geographiclib.Spheroid,
	point s2.Point,
	edge s2.Edge,
	stopAfterLessThan float64,
) float64 {
	minDistance = math.Min(minDistance, spheroidInverse(s, point, edge.V1))
	if minDistance < stopAfterLessThan {
		return minDistance
	}
	if closestPoint, ok := maybeClosestPointToEdge(edge, point); ok {
		minDistance = math.Min(minDistance, spheroidInverse(s, point, closestPoint))
		if minDistance < stopAfterLessThan {
			return minDistance
		}
	}
	return minDistance
}

// minSpheroidDistancePointToPolyline returns the minimum distance between a point
// and a polyline on a given spheroid.
func minSpheroidDistancePointToPolyline(
	s geographiclib.Spheroid, a s2.Point, b *s2.Polyline, stopAfterLessThan float64,
) float64 {
	minDistance := spheroidInverse(s, a, (*b)[0])
	if minDistance < stopAfterLessThan {
		return minDistance
	}
	for edgeIdx := 0; edgeIdx < b.NumEdges(); edgeIdx++ {
		edge := b.Edge(edgeIdx)
		minDistance = minSpheroidDistancePointToEdgeEndAndProjection(
			minDistance,
			s,
			a,
			edge,
			stopAfterLessThan,
		)
		if minDistance < stopAfterLessThan {
			return minDistance
		}
	}
	return minDistance
}

// minSpheroidDistancePointToPolygon returns the minimum distance between a point
// and a polyline on a given spheroid.
func minSpheroidDistancePointToPolygon(
	s geographiclib.Spheroid, a s2.Point, b *s2.Polygon, stopAfterLessThan float64,
) float64 {
	if b.ContainsPoint(a) {
		return 0
	}
	minDistance := math.MaxFloat64
	for _, loop := range b.Loops() {
		// Note we don't have to check min distance against the 0th point, because
		// we always check the last point (which is the same as the first point!)
		for edgeIdx := 0; edgeIdx < loop.NumEdges(); edgeIdx++ {
			edge := loop.Edge(edgeIdx)
			minDistance = minSpheroidDistancePointToEdgeEndAndProjection(
				minDistance,
				s,
				a,
				edge,
				stopAfterLessThan,
			)
			if minDistance < stopAfterLessThan {
				return minDistance
			}
		}
	}
	return minDistance
}

// minDistanceSphere calculates the minimum distance between anything
// in aRegions anything in bRegions on the spheroid.
// It will terminate if it finds anything less than "stopAfterLessThan"
// (which is useful for DWithin), which may not give the actual minimum
// distance. Use 0.
// NOTE(otan): If s2 exposed the closest point in their min distance
// edge finder, we wouldn't have to roll this ourselves. But here I stand.
func minDistanceSpheroid(
	aRegions []s2.Region, bRegions []s2.Region, stopAfterLessThan float64,
) (float64, error) {
	minDistance := math.MaxFloat64
	spheroid := geographiclib.WGS84Spheroid
	for _, aRegion := range aRegions {
		for _, bRegion := range bRegions {
			switch aRegion := aRegion.(type) {
			case s2.Point:
				switch bRegion := bRegion.(type) {
				case s2.Point:
					minDistance = math.Min(minDistance, spheroidInverse(spheroid, aRegion, bRegion))
				case *s2.Polyline:
					minDistance = math.Min(
						minDistance,
						minSpheroidDistancePointToPolyline(spheroid, aRegion, bRegion, stopAfterLessThan),
					)
				case *s2.Polygon:
					minDistance = math.Min(
						minDistance,
						minSpheroidDistancePointToPolygon(spheroid, aRegion, bRegion, stopAfterLessThan),
					)
				default:
					return 0, fmt.Errorf("unknown s2 type of b: %#v", bRegion)
				}
			case *s2.Polyline:
				switch bRegion := bRegion.(type) {
				case s2.Point:
					minDistance = math.Min(
						minDistance,
						minSpheroidDistancePointToPolyline(spheroid, bRegion, aRegion, stopAfterLessThan),
					)
				case *s2.Polyline:
				case *s2.Polygon:
				default:
					return 0, fmt.Errorf("unknown s2 type of b: %#v", bRegion)
				}
			case *s2.Polygon:
				switch bRegion := bRegion.(type) {
				case s2.Point:
					minDistance = math.Min(
						minDistance,
						minSpheroidDistancePointToPolygon(spheroid, bRegion, aRegion, stopAfterLessThan),
					)
				case *s2.Polyline:
				case *s2.Polygon:
				default:
					return 0, fmt.Errorf("unknown s2 type of b: %#v", bRegion)
				}
			default:
				return 0, fmt.Errorf("unknown s2 type of a: %#v", aRegion)
			}

			if minDistance < stopAfterLessThan {
				return minDistance, nil
			}
		}
	}
	return minDistance, nil
}

// minDistanceSphere calculates the minimum distance between anything
// in aRegions anything in bRegions on the sphere.
// This is delegated to S2, which has better accuracy and speed than
// the spheroid implementation.
// NOTE: if you need an early exit point, use `withinDistanceSphere` (TODO(otan): implement).
func minDistanceSphere(aRegions []s2.Region, bRegions []s2.Region) (float64, error) {
	// S2 has `NewClosestEdgeQuery`, which only works between edges.
	// This is optimized by using the ShapeIndex for regions with edges.
	// As such, we have to file everything into "POINT" and "LINESTRING/POLYGON".
	// We will stuff all edges into aShapeIndex and bShapeIndex, which we will use
	// to compare against each other as well as against the points of each other.
	aShapeIndex, aPoints, err := s2RegionsToPointsAndShapeIndexes(aRegions)
	if err != nil {
		return 0, err
	}
	bShapeIndex, bPoints, err := s2RegionsToPointsAndShapeIndexes(bRegions)
	if err != nil {
		return 0, err
	}
	opts := s2.NewClosestEdgeQueryOptions()

	minD := s1.StraightChordAngle

	// Compare aShapeIndex to bShapeIndex as well as all points in B.
	if aShapeIndex.Len() > 0 {
		aQuery := s2.NewClosestEdgeQuery(aShapeIndex, opts)
		if bShapeIndex.Len() > 0 {
			minD = minChordAngle(minD, aQuery.Distance(s2.NewMinDistanceToShapeIndexTarget(bShapeIndex)))
			if minD == 0 {
				return 0, nil
			}
		}
		for _, bPoint := range bPoints {
			minD = minChordAngle(minD, aQuery.Distance(s2.NewMinDistanceToPointTarget(bPoint)))
			if minD == 0 {
				return 0, nil
			}
		}
	}

	// Then try compare all points against bShapeIndex, then any points in B.
	bQuery := s2.NewClosestEdgeQuery(bShapeIndex, opts)
	for _, aPoint := range aPoints {
		if bShapeIndex.Len() > 0 {
			minD = minChordAngle(minD, bQuery.Distance(s2.NewMinDistanceToPointTarget(aPoint)))
			if minD == 0 {
				return 0, nil
			}
		}
		for _, bPoint := range bPoints {
			minD = minChordAngle(minD, s2.ChordAngleBetweenPoints(aPoint, bPoint))
			if minD == 0 {
				return 0, nil
			}
		}
	}
	return minD.Angle().Radians() * wgs84SphereRadius, nil
}
