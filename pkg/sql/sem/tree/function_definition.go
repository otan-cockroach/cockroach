// Copyright 2016 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tree

import "github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"

// FunctionDefinition implements a reference to the (possibly several)
// overloads for a built-in function.
type FunctionDefinition struct {
	// Name is the short name of the function.
	Name string

	// Definition is the set of overloads for this function name.
	// We use []overloadImpl here although all the uses of this struct
	// could actually write a []Overload, because we want to share
	// the code with typeCheckOverloadedExprs().
	Definition []overloadImpl

	// FunctionProperties are the properties common to all overloads.
	FunctionProperties
}

// Volatility indicates whether the result of a function is dependent *only*
// on the values of its explicit arguments, or can change due to outside factors
// (such as parameter variables or table contents).
//
// NOTE: functions having side-effects, such as setval(),
// must be labeled volatile to ensure they will not get optimized away,
// even if the actual return value is not changeable.
//
// This matches the PostgreSQL definition of `provolatile` in `pg_proc`.
type Volatility byte

const (
	// VolatilityImmutable means the operator cannot modify the database, the transaction state,
	// or any other state. It cannot depend on configuration settings and is
	// guaranteed to return the same results given the same arguments forever.
	// Immutable operators can be constant folded.
	// Examples: log, from_json.
	VolatilityImmutable Volatility = 'i'
	// VolatilityStable means the operator cannot modify the database or the transaction state
	// and is guaranteed to return the same results given the same arguments
	// whenever it is evaluated within the same statement. Multiple calls to a
	// stable operator can be optimized to a single call.
	// Examples: current_timestamp, current_date.
	VolatilityStable Volatility = 's'
	// VolatilityVolatile means the operator can do anything, including modifying database state.
	// Examples: random, crdb_internal.force_error, nextval.
	VolatilityVolatile Volatility = 'v'
)

// FunctionProperties defines the properties of the built-in
// functions that are common across all overloads.
type FunctionProperties struct {
	// UnsupportedWithIssue, if non-zero indicates the built-in is not
	// really supported; the name is a placeholder. Value -1 just says
	// "not supported" without an issue to link; values > 0 provide an
	// issue number to link.
	UnsupportedWithIssue int

	// NullableArgs is set to true when a function's definition can
	// handle NULL arguments. When set, the function will be given the
	// chance to see NULL arguments. When not, the function will
	// evaluate directly to NULL in the presence of any NULL arguments.
	//
	// NOTE: when set, a function should be prepared for any of its arguments to
	// be NULL and should act accordingly.
	NullableArgs bool

	// Private, when set to true, indicates the built-in function is not
	// available for use by user queries. This is currently used by some
	// aggregates due to issue #10495.
	Private bool

	// NeedsRepeatedEvaluation is set to true when a function may change
	// at every row whether or not it is applied to an expression that
	// contains row-dependent variables. Used e.g. by `random` and
	// aggregate functions.
	NeedsRepeatedEvaluation bool

	// DistsqlBlacklist is set to true when a function depends on
	// members of the EvalContext that are not marshaled by DistSQL
	// (e.g. planner). Currently used for DistSQL to determine if
	// expressions can be evaluated on a different node without sending
	// over the EvalContext.
	//
	// TODO(andrei): Get rid of the planner from the EvalContext and then we can
	// get rid of this blacklist.
	DistsqlBlacklist bool

	// Class is the kind of built-in function (normal/aggregate/window/etc.)
	Class FunctionClass

	// Category is used to generate documentation strings.
	Category string

	// ReturnLabels can be used to override the return column name of a
	// function in a FROM clause.
	// This satisfies a Postgres quirk where some json functions have
	// different return labels when used in SELECT or FROM clause.
	ReturnLabels []string

	// AmbiguousReturnType is true if the builtin's return type can't be
	// determined without extra context. This is used for formatting builtins
	// with the FmtParsable directive.
	AmbiguousReturnType bool
}

// FunctionClass specifies the class of the builtin function.
type FunctionClass int

const (
	// NormalClass is a standard builtin function.
	NormalClass FunctionClass = iota
	// AggregateClass is a builtin aggregate function.
	AggregateClass
	// WindowClass is a builtin window function.
	WindowClass
	// GeneratorClass is a builtin generator function.
	GeneratorClass
)

// Avoid vet warning about unused enum value.
var _ = NormalClass

// NewFunctionDefinition allocates a function definition corresponding
// to the given built-in definition.
func NewFunctionDefinition(
	name string, props *FunctionProperties, def []Overload,
) *FunctionDefinition {
	overloads := make([]overloadImpl, len(def))

	for i := range def {
		if def[i].PreferredOverload {
			// Builtins with a preferred overload are always ambiguous.
			props.AmbiguousReturnType = true
		}
		// Produce separate telemetry for each overload.
		def[i].counter = sqltelemetry.BuiltinCounter(name, def[i].Signature(false))

		overloads[i] = &def[i]
	}
	return &FunctionDefinition{
		Name:               name,
		Definition:         overloads,
		FunctionProperties: *props,
	}
}

// FunDefs holds pre-allocated FunctionDefinition instances
// for every builtin function. Initialized by builtins.init().
var FunDefs map[string]*FunctionDefinition

// Format implements the NodeFormatter interface.
func (fd *FunctionDefinition) Format(ctx *FmtCtx) {
	ctx.WriteString(fd.Name)
}
func (fd *FunctionDefinition) String() string { return AsString(fd) }
