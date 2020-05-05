// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package geomfn contains functions that are used for geometry-based builtins.
package geomfn

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/errors"
	"github.com/golang/geo/r3"
	"github.com/twpayne/go-geom"
)

// point represents a point in the 2D space.
type point struct {
	r3.Vector
}

// pointFromGeomPoint initializes a point from a geom.Point.
func pointFromGeomPoint(p *geom.Point) point {
	return point{r3.Vector{X: p.X(), Y: p.Y()}}
}

// pointFromGeomCoord initializes a point from a geom.Point.
func pointFromGeomCoord(p geom.Coord) point {
	return point{r3.Vector{X: p.X(), Y: p.Y()}}
}

// inBoundingCross returns whether a point is between a given edge.
// This includes the boundary.
// It only checks whether X _or_ Y is within the range of values covered by the edge.
func (p *point) inBoundingCross(e edge) bool {
	return (e.V0.X <= p.X && p.X < e.V1.X) || (e.V1.X <= p.X && p.X < e.V0.X) ||
		(e.V0.Y <= p.Y && p.Y < e.V1.Y) || (e.V1.Y <= p.Y && p.Y < e.V0.Y)
}

// edge represents an edge between two points in the 2D space.
type edge struct {
	V0 point
	V1 point
}

// maybeFoundClosestPoint projects a point p onto the edge line.
// It returns and a bool representing whether the point is on the edge.
// If true, it will also return a point that represents the closest point to the edge.
func (e *edge) maybeFindClosestPoint(p point) (point, bool) {
	// From http://www.faqs.org/faqs/graphics/algorithms-faq/, section 1.02
	//
	//  Let the point be C (Cx,Cy) and the line be AB (Ax,Ay) to (Bx,By).
	//  Let P be the point of perpendicular projection of C on AB.  The parameter
	//  r, which indicates P's position along AB, is computed by the dot product
	//  of AC and AB divided by the square of the length of AB:
	//
	//  (1)     AC dot AB
	//      r = ---------
	//          ||AB||^2
	//
	//  r has the following meaning:
	//
	//      r=0      P = A
	//      r=1      P = B
	//      r<0      P is on the backward extension of AB
	//      r>1      P is on the forward extension of AB
	//      0<r<1    P is interior to AB
	if p == e.V0 {
		return p, true
	}
	if p == e.V1 {
		return p, true
	}

	ac := p.Vector.Sub(e.V0.Vector)
	ab := e.V1.Vector.Sub(e.V0.Vector)

	r := ac.Dot(ab) / ab.Norm2()
	if r < 0 || r > 1 {
		return point{}, false
	}
	return point{Vector: e.V0.Add(e.V1.Vector.Mul(r))}, true
}

// side corresponds to the side in which a point is relative to a line.
type pointSide int

const (
	pointSideLeft  pointSide = -1
	pointSideOn    pointSide = 0
	pointSideRight pointSide = 1
)

// findPointSide finds which side a point is relative to the infinite line
// given by the edge.
// Note this side is relative to the orientation of the line.
func findPointSide(p point, e edge) pointSide {
	// This is the equivalent of using the point-gradient formula
	// and determining the sign, i.e. the sign of
	// d = (x-x1)(y2-y1) - (y-y1)(x2-x1)
	// where (x1,y1) and (x2,y2) is the edge and (x,y) is the point
	sign := (p.X-e.V0.X)*(e.V1.Y-e.V0.Y) - (e.V1.X-e.V0.X)*(p.Y-e.V0.Y)
	switch {
	case sign == 0:
		return pointSideOn
	case sign > 0:
		return pointSideRight
	default:
		return pointSideLeft
	}
}

// flattenGeometry flattens a geo.Geometry object.
func flattenGeometry(g *geo.Geometry) ([]geom.T, error) {
	f, err := g.AsGeomT()
	if err != nil {
		return nil, err
	}
	return flattenGeomT(f)
}

// flattenGeomT decomposes geom.T collections to individual geom.T components.
func flattenGeomT(g geom.T) ([]geom.T, error) {
	switch g := g.(type) {
	case *geom.Point:
		return []geom.T{g}, nil
	case *geom.LineString:
		return []geom.T{g}, nil
	case *geom.Polygon:
		return []geom.T{g}, nil
	case *geom.MultiPoint:
		ret := make([]geom.T, g.NumPoints())
		for i := 0; i < g.NumPoints(); i++ {
			ret[i] = g.Point(i)
		}
		return ret, nil
	case *geom.MultiLineString:
		ret := make([]geom.T, g.NumLineStrings())
		for i := 0; i < g.NumLineStrings(); i++ {
			ret[i] = g.LineString(i)
		}
		return ret, nil
	case *geom.MultiPolygon:
		ret := make([]geom.T, g.NumPolygons())
		for i := 0; i < g.NumPolygons(); i++ {
			ret[i] = g.Polygon(i)
		}
		return ret, nil
	case *geom.GeometryCollection:
		ret := make([]geom.T, g.NumGeoms())
		for i := 0; i < g.NumGeoms(); i++ {
			ret[i] = g.Geom(i)
		}
		return ret, nil
	}
	return nil, errors.Newf("unknown geom: %T", g)
}
