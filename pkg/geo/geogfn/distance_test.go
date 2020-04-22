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
	"math"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/stretchr/testify/require"
)

func TestDistance(t *testing.T) {
	testCases := []struct {
		desc                     string
		a                        string
		b                        string
		expectedSphereDistance   float64
		expectedSpheroidDistance float64
	}{
		{
			"POINT to itself",
			"POINT(1.0 1.0)",
			"POINT(1.0 1.0)",
			0,
			0,
		},
		{
			"POINT to POINT (CDG to LAX)",
			"POINT(-118.4079 33.9434)",
			"POINT(2.5559 49.0083)",
			9.103087983009e+06,
			9.124665273176e+06,
		},
		{
			"LINESTRING to POINT where POINT is on vertex",
			"LINESTRING(2.0 2.0, 3.0 3.0)",
			"POINT(3.0 3.0)",
			0,
			0,
		},
		{
			"LINESTRING to POINT where POINT is closest to a vertex",
			"LINESTRING(1.0 1.0, 2.0 2.0, 3.0 3.0)",
			"POINT(5.0 5.0)",
			314116.2410064,
			313424.6522007,
		},
		{
			"LINESTRING to POINT where POINT is closer than the edge vertices",
			"LINESTRING(1.0 1.0, 2.0 2.0, 3.0 3.0)",
			"POINT(2.4 2.6)",
			15695.12116722,
			15660.43959933,
		},
		{
			"LINESTRING to POINT on the line",
			"LINESTRING(0.0 0.0, 1.0 1.0, 2.0 2.0, 3.0 3.0)",
			"POINT(1.5 1.5001714)",
			0,
			0,
		},
		{
			"POLYGON to POINT inside the polygon",
			"POLYGON((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0))",
			"POINT(0.5 0.5)",
			0,
			0,
		},
		{
			"POLYGON to POINT on vertex of the polygon",
			"POLYGON((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0))",
			"POINT(1.0 1.0)",
			0,
			0,
		},
		{
			"POLYGON to POINT outside the polygon",
			"POLYGON((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0))",
			"POINT(1.5 1.6)",
			86836.81014284,
			86591.2400406,
		},
		{},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			a, err := geo.ParseGeography(tc.a)
			require.NoError(t, err)
			b, err := geo.ParseGeography(tc.b)
			require.NoError(t, err)

			// Allow a 1cm margin of error for results.
			t.Run("sphere", func(t *testing.T) {
				distance, err := Distance(a, b, UseSphere)
				require.NoError(t, err)
				require.LessOrEqualf(
					t,
					math.Abs(tc.expectedSphereDistance-distance),
					0.01,
					"expected %f, got %f",
					tc.expectedSphereDistance,
					distance,
				)
				distance, err = Distance(b, a, UseSphere)
				require.NoError(t, err)
				require.LessOrEqualf(
					t,
					math.Abs(tc.expectedSphereDistance-distance),
					0.01,
					"expected %f, got %f",
					tc.expectedSphereDistance,
					distance,
				)
			})
			t.Run("spheroid", func(t *testing.T) {
				distance, err := Distance(a, b, UseSpheroid)
				require.NoError(t, err)
				require.LessOrEqualf(
					t,
					math.Abs(tc.expectedSpheroidDistance-distance),
					0.01,
					"expected %f, got %f",
					tc.expectedSpheroidDistance,
					distance,
				)
				distance, err = Distance(b, a, UseSpheroid)
				require.NoError(t, err)
				require.LessOrEqualf(
					t,
					math.Abs(tc.expectedSpheroidDistance-distance),
					0.01,
					"expected %f, got %f",
					tc.expectedSpheroidDistance,
					distance,
				)
			})
		})
	}
}
