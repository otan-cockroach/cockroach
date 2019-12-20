// Copyright 2017 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package timeutil

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

const (
	fixedOffsetPrefix = "fixed offset:"
	offsetBoundSecs   = 167*60*60 + 59*60
)

var timezoneOffsetRegex = regexp.MustCompile(`(?i)(GMT|UTC)([+-])(\d{1,3}(:[0-5]?\d){0,2})\b$`)

// FixedOffsetTimeZoneToLocation creates a time.Location with a set offset and
// with a name that can be marshaled by crdb between nodes.
func FixedOffsetTimeZoneToLocation(offset int, origRepr string) *time.Location {
	return time.FixedZone(
		fmt.Sprintf("%s%d (%s)", fixedOffsetPrefix, offset, origRepr),
		offset)
}

// ParseStringAsTimeZone attempts to parse a string as a timezone,
// returning the location if successful, and a bool about whether
// the conversion was successful.
func ParseStringAsTimeZone(str string) (*time.Location, bool) {
	str = strings.TrimSpace(str)

	offset, origRepr, parsed := parseFixedOffsetTimeZone(str)
	if parsed {
		return FixedOffsetTimeZoneToLocation(offset, origRepr), true
	}

	locTransforms := []func(string) string{
		func(s string) string { return s },
		strings.ToUpper,
		strings.ToTitle,
	}
	for _, transform := range locTransforms {
		if loc, err := LoadLocation(transform(str)); err == nil {
			return loc, true
		}
	}

	// Maybe the string is coming from pgwire as a simple number.
	intVal, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return timeutil.FixedOffsetTimeZoneToLocation(int(intVal)*60*60, s)
	}

	strOffset, ok := timeZoneOffsetStringConversion(str)
	return FixedOffsetTimeZoneToLocation(int(strOffset), str), ok
}

// parseFixedOffsetTimeZone takes the string representation of a time.Location
// created by FixedOffsetTimeZoneToLocation and parses it to the offset and the
// original representation specified by the user. The bool returned is true if
// parsing was successful.
//
// The strings produced by FixedOffsetTimeZoneToLocation look like
// "<fixedOffsetPrefix><offset> (<origRepr>)".
// TODO(#42404): this is not the format given by the results in
// pgwire/testdata/connection_params.
func parseFixedOffsetTimeZone(location string) (offset int, origRepr string, success bool) {
	if !strings.HasPrefix(location, fixedOffsetPrefix) {
		return 0, "", false
	}
	location = strings.TrimPrefix(location, fixedOffsetPrefix)
	parts := strings.SplitN(location, " ", 2)
	if len(parts) < 2 {
		return 0, "", false
	}

	offset, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", false
	}

	origRepr = parts[1]
	if !strings.HasPrefix(origRepr, "(") || !strings.HasSuffix(origRepr, ")") {
		return 0, "", false
	}
	return offset, strings.TrimSuffix(strings.TrimPrefix(origRepr, "("), ")"), true
}

// timeZoneOffsetStringConversion converts a time string to offset seconds.
// Supported time zone strings: GMT/UTC±[00:00:00 - 169:59:00].
// Seconds/minutes omittable and is case insensitive.
func timeZoneOffsetStringConversion(s string) (offset int64, ok bool) {
	submatch := timezoneOffsetRegex.FindStringSubmatch(strings.ReplaceAll(s, " ", ""))
	if len(submatch) == 0 {
		return 0, false
	}
	prefix := string(submatch[0][3])
	timeString := string(submatch[3])

	var (
		hoursString   = "0"
		minutesString = "0"
		secondsString = "0"
	)
	if strings.Contains(timeString, ":") {
		offsets := strings.Split(timeString, ":")
		hoursString, minutesString = offsets[0], offsets[1]
		if len(offsets) == 3 {
			secondsString = offsets[2]
		}
	} else {
		hoursString = timeString
	}

	hours, _ := strconv.ParseInt(hoursString, 10, 64)
	minutes, _ := strconv.ParseInt(minutesString, 10, 64)
	seconds, _ := strconv.ParseInt(secondsString, 10, 64)
	offset = (hours * 60 * 60) + (minutes * 60) + seconds

	if prefix == "-" {
		offset *= -1
	}

	if offset > offsetBoundSecs || offset < -offsetBoundSecs {
		return 0, false
	}
	return offset, true
}
