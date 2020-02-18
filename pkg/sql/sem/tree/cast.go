// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tree

import "github.com/lib/pq/oid"

// CastContext specifies how a given Castable can be used.
type CastContext uint8

// CastMethod specifies how the cast is done.
// This is legacy from postgres, but affects type overloads,
// so we also include it.
type CastMethod uint8

const (
	// CastContextImplicit implies the cast can be set implicitly
	// in any context.
	CastContextImplicit CastContext = iota
	// CastContextAssign implies that the cast can done implicitly
	// in assign contexts (e.g. on UPDATE and INSERT).
	CastContextAssign
	// CastContextExplicit implies that the cast can only be used
	// in an explicit context.
	CastContextExplicit
)

// Castable defines how a method can be cast.
type Castable struct {
	Source  oid.Oid
	Target  oid.Oid
	Context CastContext
	Method  CastMethod
	// TODO: move eval.go's PerformCast to here.
}

// CastableMap defines which oids can be cast to which other oids.
// The map goes from src oid -> target oid -> Castable.
// Source and Target are initialized in init().
var CastableMap = map[oid.Oid]map[oid.Oid]Castable{
	oid.T_timestamp: map[oid.Oid]Castable{
		oid.T_timestamptz: {
			Context: CastContextImplicit,
		},
	},
	oid.T_timestamptz: map[oid.Oid]Castable{
		oid.T_timestamp: {
			Context: CastContextAssign,
		},
	},
}

func init() {
	for src, tgts := range CastableMap {
		for tgt := range tgts {
			ent := CastableMap[src][tgt]
			ent.Source = src
			ent.Target = tgt
			CastableMap[src][tgt] = ent
		}
	}
}

// FindCast returns whether the given src oid can be cast to
// the target oid, given the CastContext.
// Returns a Castable object if it can be cast, and a bool whether the cast is valid.
func FindCast(src oid.Oid, tgt oid.Oid, ctx CastContext) (Castable, bool) {
	if tgts, ok := CastableMap[src]; ok {
		if castable, ok := tgts[tgt]; ok {
			return castable, castable.Context >= ctx
		}
	}
	return Castable{}, false
}
