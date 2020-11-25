// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package descpb

// Region is an alias for a region stored on the database.
type Region string

// ConstraintKey returns the key to use for constraints for a given region
// in the CONFIGURE ZONE syntax.
func (r Region) ConstraintKey() string {
	return "+region=" + string(r)
}
