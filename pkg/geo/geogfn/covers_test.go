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
	"testing"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/stretchr/testify/require"
)

func TestCovers(t *testing.T) {
	testCases := []struct {
		desc     string
		a        string
		b        string
		expected bool
	}{
		{
			"Point covers the same Point",
			"POINT(1.0 1.0)",
			"POINT(1.0 1.0)",
			true,
		},
		{
			"Point does not cover different Point",
			"POINT(1.0 1.0)",
			"POINT(1.0 1.1)",
			false,
		},
		{
			"Point does not cover different LineString",
			"POINT(1.0 1.0)",
			"LINESTRING(1.0 1.0, 2.0 2.0, 3.0 3.0)",
			false,
		},
		{
			"Point does not cover different Polygon",
			"POINT(1.0 1.0)",
			"POLYGON((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0))",
			false,
		},
		{
			"LineString covers Point at the start",
			"LINESTRING(1.0 1.0, 2.0 2.0, 3.0 3.0)",
			"POINT(1.0 1.0)",
			true,
		},
		{
			"LineString covers Point at the middle",
			"LINESTRING(1.0 1.0, 2.0 2.0, 3.0 3.0)",
			"POINT(2.0 2.0)",
			true,
		},
		{
			"LineString covers Point at the end",
			"LINESTRING(1.0 1.0, 2.0 2.0, 3.0 3.0)",
			"POINT(3.0 3.0)",
			true,
		},
		{
			"LineString covers Point in between",
			"LINESTRING(1.0 1.0, 2.0 2.0, 3.0 3.0)",
			"POINT(1.5 1.5001714)",
			true,
		},
		{
			"LineString does not cover any Point out of the way",
			"LINESTRING(1.0 1.0, 2.0 2.0, 3.0 3.0)",
			"POINT(2.0 4.0)",
			false,
		},
		{
			"Linestring covers itself",
			"LINESTRING(1.0 1.0, 2.0 2.0, 3.0 3.0)",
			"LINESTRING(1.0 1.0, 2.0 2.0, 3.0 3.0)",
			true,
		},
		{
			"Linestring covers a substring of itself",
			"LINESTRING(1.0 1.0, 2.0 2.0, 3.0 3.0)",
			"LINESTRING(1.0 1.0, 2.0 2.0)",
			true,
		},
		{
			"Linestring covers a substring of itself",
			"LINESTRING(1.0 1.0, 2.0 2.0, 3.0 3.0)",
			"LINESTRING(2.0 2.0, 3.0 3.0)",
			true,
		},
		{
			"Linestring covers a substring of itself",
			"LINESTRING(1.0 1.0, 2.0 2.0, 3.0 3.0, 4.0 4.0)",
			"LINESTRING(2.0 2.0, 3.0 3.0)",
			true,
		},
		{
			"LineString does not cover any Polygon",
			"POINT(1.0 1.0)",
			"POLYGON((0.0 0.0, 1.0 0.0, 1.0 1.0, 0.0 1.0, 0.0 0.0))",
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			fmt.Printf("%s\n", tc.desc)
			a, err := geo.ParseGeography(tc.a)
			require.NoError(t, err)
			b, err := geo.ParseGeography(tc.b)
			require.NoError(t, err)

			covers, err := Covers(a, b)
			require.NoError(t, err)
			require.Equal(t, tc.expected, covers)

			coveredBy, err := CoveredBy(b, a)
			require.NoError(t, err)
			require.Equal(t, tc.expected, coveredBy)
		})
	}
}
