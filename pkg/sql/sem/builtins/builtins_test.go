// Copyright 2016 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package builtins

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategory(t *testing.T) {
	if expected, actual := categoryString, builtins["lower"].props.Category; expected != actual {
		t.Fatalf("bad category: expected %q got %q", expected, actual)
	}
	if expected, actual := categoryString, builtins["length"].props.Category; expected != actual {
		t.Fatalf("bad category: expected %q got %q", expected, actual)
	}
	if expected, actual := categoryDateAndTime, builtins["now"].props.Category; expected != actual {
		t.Fatalf("bad category: expected %q got %q", expected, actual)
	}
	if expected, actual := categorySystemInfo, builtins["version"].props.Category; expected != actual {
		t.Fatalf("bad category: expected %q got %q", expected, actual)
	}
}

// TestGenerateUniqueIDOrder verifies the expected ordering of
// GenerateUniqueID.
func TestGenerateUniqueIDOrder(t *testing.T) {
	tests := []tree.DInt{
		GenerateUniqueID(0, 0),
		GenerateUniqueID(1, 0),
		GenerateUniqueID(2<<15, 0),
		GenerateUniqueID(0, 1),
		GenerateUniqueID(0, 10000),
		GenerateUniqueInt(0),
	}
	prev := tests[0]
	for _, tc := range tests[1:] {
		if tc <= prev {
			t.Fatalf("%d > %d", tc, prev)
		}
	}
}

func TestStringToArrayAndBack(t *testing.T) {
	// s allows us to have a string pointer literal.
	s := func(x string) *string { return &x }
	fs := func(x *string) string {
		if x != nil {
			return *x
		}
		return "<nil>"
	}
	cases := []struct {
		input    string
		sep      *string
		nullStr  *string
		expected []*string
	}{
		{`abcxdef`, s(`x`), nil, []*string{s(`abc`), s(`def`)}},
		{`xxx`, s(`x`), nil, []*string{s(``), s(``), s(``), s(``)}},
		{`xxx`, s(`xx`), nil, []*string{s(``), s(`x`)}},
		{`abcxdef`, s(``), nil, []*string{s(`abcxdef`)}},
		{`abcxdef`, s(`abcxdef`), nil, []*string{s(``), s(``)}},
		{`abcxdef`, s(`x`), s(`abc`), []*string{nil, s(`def`)}},
		{`abcxdef`, s(`x`), s(`x`), []*string{s(`abc`), s(`def`)}},
		{`abcxdef`, s(`x`), s(``), []*string{s(`abc`), s(`def`)}},
		{``, s(`x`), s(``), []*string{}},
		{``, s(``), s(``), []*string{}},
		{``, s(`x`), nil, []*string{}},
		{``, s(``), nil, []*string{}},
		{`abcxdef`, nil, nil, []*string{s(`a`), s(`b`), s(`c`), s(`x`), s(`d`), s(`e`), s(`f`)}},
		{`abcxdef`, nil, s(`abc`), []*string{s(`a`), s(`b`), s(`c`), s(`x`), s(`d`), s(`e`), s(`f`)}},
		{`abcxdef`, nil, s(`x`), []*string{s(`a`), s(`b`), s(`c`), nil, s(`d`), s(`e`), s(`f`)}},
		{`abcxdef`, nil, s(``), []*string{s(`a`), s(`b`), s(`c`), s(`x`), s(`d`), s(`e`), s(`f`)}},
		{``, nil, s(``), []*string{}},
		{``, nil, nil, []*string{}},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("string_to_array(%q, %q)", tc.input, fs(tc.sep)), func(t *testing.T) {
			result, err := stringToArray(tc.input, tc.sep, tc.nullStr)
			if err != nil {
				t.Fatal(err)
			}

			expectedArray := tree.NewDArray(types.String)
			for _, s := range tc.expected {
				datum := tree.DNull
				if s != nil {
					datum = tree.NewDString(*s)
				}
				if err := expectedArray.Append(datum); err != nil {
					t.Fatal(err)
				}
			}

			evalContext := tree.NewTestingEvalContext(cluster.MakeTestingClusterSettings())
			if result.Compare(evalContext, expectedArray) != 0 {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}

			if tc.sep == nil {
				return
			}

			s, err := arrayToString(result.(*tree.DArray), *tc.sep, tc.nullStr)
			if err != nil {
				t.Fatal(err)
			}
			if s == tree.DNull {
				t.Errorf("expected not null, found null")
			}

			ds := s.(*tree.DString)
			fmt.Println(ds)
			if string(*ds) != tc.input {
				t.Errorf("original %s, roundtripped %s", tc.input, s)
			}
		})
	}
}

func TestEscapeFormat(t *testing.T) {
	testCases := []struct {
		bytes []byte
		str   string
	}{
		{[]byte{}, ``},
		{[]byte{'a', 'b', 'c'}, `abc`},
		{[]byte{'a', 'b', 'c', 'd'}, `abcd`},
		{[]byte{'a', 'b', 0, 'd'}, `ab\000d`},
		{[]byte{'a', 'b', 0, 0, 'd'}, `ab\000\000d`},
		{[]byte{'a', 'b', 0, 'a', 'b', 'c', 0, 'd'}, `ab\000abc\000d`},
		{[]byte{'a', 'b', 0, 0}, `ab\000\000`},
		{[]byte{'a', 'b', '\\', 'd'}, `ab\\d`},
		{[]byte{'a', 'b', 200, 'd'}, `ab\310d`},
		{[]byte{'a', 'b', 7, 'd'}, "ab\x07d"},
	}

	for _, tc := range testCases {
		t.Run(tc.str, func(t *testing.T) {
			result := encodeEscape(tc.bytes)
			if result != tc.str {
				t.Fatalf("expected %q, got %q", tc.str, result)
			}

			decodedResult, err := decodeEscape(tc.str)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(decodedResult, tc.bytes) {
				t.Fatalf("expected %q, got %#v", tc.bytes, decodedResult)
			}
		})
	}
}

func TestEscapeFormatRandom(t *testing.T) {
	for i := 0; i < 1000; i++ {
		b := make([]byte, rand.Intn(100))
		for j := 0; j < len(b); j++ {
			b[j] = byte(rand.Intn(256))
		}
		str := encodeEscape(b)
		decodedResult, err := decodeEscape(str)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(decodedResult, b) {
			t.Fatalf("generated %#v, after round-tripping got %#v", b, decodedResult)
		}
	}
}

func TestLPadRPad(t *testing.T) {
	testCases := []struct {
		padFn    func(string, int, string) (string, error)
		str      string
		length   int
		fill     string
		expected string
	}{
		{lpad, "abc", 1, "xy", "a"},
		{lpad, "abc", 2, "xy", "ab"},
		{lpad, "abc", 3, "xy", "abc"},
		{lpad, "abc", 5, "xy", "xyabc"},
		{lpad, "abc", 6, "xy", "xyxabc"},
		{lpad, "abc", 7, "xy", "xyxyabc"},
		{lpad, "abc", 1, " ", "a"},
		{lpad, "abc", 2, " ", "ab"},
		{lpad, "abc", 3, " ", "abc"},
		{lpad, "abc", 5, " ", "  abc"},
		{lpad, "Hello, 世界", 9, " ", "Hello, 世界"},
		{lpad, "Hello, 世界", 10, " ", " Hello, 世界"},
		{lpad, "Hello", 8, "世界", "世界世Hello"},
		{lpad, "foo", -1, "世界", ""},
		{rpad, "abc", 1, "xy", "a"},
		{rpad, "abc", 2, "xy", "ab"},
		{rpad, "abc", 3, "xy", "abc"},
		{rpad, "abc", 5, "xy", "abcxy"},
		{rpad, "abc", 6, "xy", "abcxyx"},
		{rpad, "abc", 7, "xy", "abcxyxy"},
		{rpad, "abc", 1, " ", "a"},
		{rpad, "abc", 2, " ", "ab"},
		{rpad, "abc", 3, " ", "abc"},
		{rpad, "abc", 5, " ", "abc  "},
		{rpad, "abc", 5, " ", "abc  "},
		{rpad, "Hello, 世界", 9, " ", "Hello, 世界"},
		{rpad, "Hello, 世界", 10, " ", "Hello, 世界 "},
		{rpad, "Hello", 8, "世界", "Hello世界世"},
		{rpad, "foo", -1, "世界", ""},
	}
	for _, tc := range testCases {
		out, err := tc.padFn(tc.str, tc.length, tc.fill)
		if err != nil {
			t.Errorf("Found err %v, expected nil", err)
		}
		if out != tc.expected {
			t.Errorf("expected %s, found %s", tc.expected, out)
		}
	}
}

func TestTruncateTimestamp(t *testing.T) {
	defer leaktest.AfterTest(t)()

	loc, err := timeutil.LoadLocation("Australia/Sydney")
	require.NoError(t, err)

	testCases := []struct {
		fromTime time.Time
		timeSpan string
		expected *tree.DTimestampTZ
	}{
		{time.Date(2118, time.March, 11, 5, 6, 7, 80009001, loc), "millennium", tree.MakeDTimestampTZ(time.Date(2001, time.January, 1, 0, 0, 0, 0, loc), time.Microsecond)},
		{time.Date(2118, time.March, 11, 5, 6, 7, 80009001, loc), "century", tree.MakeDTimestampTZ(time.Date(2101, time.January, 1, 0, 0, 0, 0, loc), time.Microsecond)},
		{time.Date(2118, time.March, 11, 5, 6, 7, 80009001, loc), "decade", tree.MakeDTimestampTZ(time.Date(2110, time.January, 1, 0, 0, 0, 0, loc), time.Microsecond)},
		{time.Date(2118, time.March, 11, 5, 6, 7, 80009001, loc), "year", tree.MakeDTimestampTZ(time.Date(2118, time.January, 1, 0, 0, 0, 0, loc), time.Microsecond)},
		{time.Date(2118, time.March, 11, 5, 6, 7, 80009001, loc), "quarter", tree.MakeDTimestampTZ(time.Date(2118, time.January, 1, 0, 0, 0, 0, loc), time.Microsecond)},
		{time.Date(2118, time.March, 11, 5, 6, 7, 80009001, loc), "month", tree.MakeDTimestampTZ(time.Date(2118, time.March, 1, 0, 0, 0, 0, loc), time.Microsecond)},
		{time.Date(2118, time.March, 11, 5, 6, 7, 80009001, loc), "day", tree.MakeDTimestampTZ(time.Date(2118, time.March, 11, 0, 0, 0, 0, loc), time.Microsecond)},
		{time.Date(2118, time.March, 11, 5, 6, 7, 80009001, loc), "week", tree.MakeDTimestampTZ(time.Date(2118, time.March, 6, 0, 0, 0, 0, loc), time.Microsecond)},
		{time.Date(2118, time.March, 11, 5, 6, 7, 80009001, loc), "hour", tree.MakeDTimestampTZ(time.Date(2118, time.March, 11, 5, 0, 0, 0, loc), time.Microsecond)},
		{time.Date(2118, time.March, 11, 5, 6, 7, 80009001, loc), "second", tree.MakeDTimestampTZ(time.Date(2118, time.March, 11, 5, 6, 7, 0, loc), time.Microsecond)},
		{time.Date(2118, time.March, 11, 5, 6, 7, 80009001, loc), "millisecond", tree.MakeDTimestampTZ(time.Date(2118, time.March, 11, 5, 6, 7, 80000000, loc), time.Microsecond)},
		{time.Date(2118, time.March, 11, 5, 6, 7, 80009001, loc), "microsecond", tree.MakeDTimestampTZ(time.Date(2118, time.March, 11, 5, 6, 7, 80009000, loc), time.Microsecond)},
	}

	for _, tc := range testCases {
		t.Run(tc.timeSpan, func(t *testing.T) {
			result, err := truncateTimestamp(nil, tc.fromTime, tc.timeSpan)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFloatWidthBucket(t *testing.T) {
	testCases := []struct {
		operand  float64
		b1       float64
		b2       float64
		count    int
		expected int
	}{
		{0.5, 2, 3, 5, 0},
		{8, 2, 3, 5, 6},
		{1.5, 1, 3, 2, 1},
		{5.35, 0.024, 10.06, 5, 3},
		{-3.0, -5, 5, 10, 3},
		{1, 1, 10, 2, 1},  // minimum should be inclusive
		{10, 1, 10, 2, 3}, // maximum should be exclusive
		{4, 10, 1, 4, 3},
		{11, 10, 1, 4, 0},
		{0, 10, 1, 4, 5},
	}

	for _, tc := range testCases {
		got := widthBucket(tc.operand, tc.b1, tc.b2, tc.count)
		if got != tc.expected {
			t.Errorf("expected %d, found %d", tc.expected, got)
		}
	}
}
