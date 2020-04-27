// Copyright 2020 The Cockroach Authors.
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
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geomfn"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

// infoBuilder is used to build a detailed info string that is consistent between
// geospatial data types.
type infoBuilder struct {
	info        string
	usesGEOS    bool
	canUseIndex bool
}

func (ib infoBuilder) String() string {
	var sb strings.Builder
	sb.WriteString(ib.info)
	if ib.usesGEOS {
		sb.WriteString("\n\nThis function uses the GEOS module.")
	}
	if ib.canUseIndex {
		sb.WriteString("\n\nThis function will automatically use any available index.")
	}
	return sb.String()
}

var geoBuiltins = map[string]builtinDefinition{
	//
	// Unary functions.
	//
	// TODO(rytaft, otan): replace this with a real function.
	"st_area": makeBuiltin(
		defProps(),
		tree.Overload{
			Types: tree.ArgTypes{
				{"geometry_a", types.Geometry},
			},
			ReturnType: tree.FixedReturnType(types.Float),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				return tree.NewDFloat(0), errors.Newf("ST_Area is not yet supported.")
			},
			Info: "Returns the area of geometry_a",
		},
	),

	//
	// Binary Predicates
	//
	"st_covers": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.Covers,
			infoBuilder{
				info:        "Returns true if no point in geometry_b is outside geometry_a.",
				usesGEOS:    true,
				canUseIndex: true,
			},
		),
	),
	"st_coveredby": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.CoveredBy,
			infoBuilder{
				info:        "Returns true if no point in geometry_a is outside geometry_b.",
				usesGEOS:    true,
				canUseIndex: true,
			},
		),
	),
	"st_contains": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.Contains,
			infoBuilder{
				info: "Returns true if no points of geometry_b lie in the exterior of geometry_a, " +
					"and there is at least one point in the interior of geometry_b that lies in the interior of geometry_a.",
				usesGEOS:    true,
				canUseIndex: true,
			},
		),
	),
	"st_crosses": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.Crosses,
			infoBuilder{
				info:        "Returns true if geometry_a has some - but not all - interior points in common with geometry_b.",
				usesGEOS:    true,
				canUseIndex: true,
			},
		),
	),
	"st_equals": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.Equals,
			infoBuilder{
				info: "Returns true if geometry_a is spatially equal to geometry_b, " +
					"i.e. ST_Within(geometry_a, geometry_b) = ST_Within(geometry_b, geometry_a) = true.",
				usesGEOS:    true,
				canUseIndex: true,
			},
		),
	),
	"st_intersects": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.Intersects,
			infoBuilder{
				info:        "Returns true if geometry_a shares any portion of space with geometry_b.",
				usesGEOS:    true,
				canUseIndex: true,
			},
		),
	),
	"st_overlaps": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.Overlaps,
			infoBuilder{
				info: "Returns true if geometry_a intersects but does not completely contain geometry_b, or vice versa. " +
					`"Does not completely" implies ST_Within(geometry_a, geometry_b) = ST_Within(geometry_b, geometry_a) = false.`,
				usesGEOS:    true,
				canUseIndex: true,
			},
		),
	),
	"st_touches": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.Touches,
			infoBuilder{
				info: "Returns true if the only points in common between geometry_a and geometry_b are on the boundary. " +
					"Note points do not touch other points.",
				usesGEOS:    true,
				canUseIndex: true,
			},
		),
	),
	"st_within": makeBuiltin(
		defProps(),
		geometryOverload2BinaryPredicate(
			geomfn.Within,
			infoBuilder{
				info:        "Returns true if geometry_a is completely inside geometry_b.",
				usesGEOS:    true,
				canUseIndex: true,
			},
		),
	),

	//
	// Modifiers
	//

	"addgeometrycolumn": makeBuiltin(
		tree.FunctionProperties{
			Category: categoryGeospatial,
			Impure:   true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"catalog_name", types.String},
				{"schema_name", types.String},
				{"table_name", types.String},
				{"column_name", types.String},
				{"srid", types.Int},
				{"shape", types.String},
				{"num_dimensions", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				return addGeometryColumn(
					ctx,
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					string(tree.MustBeDString(args[2])),
					string(tree.MustBeDString(args[3])),
					int(tree.MustBeDInt(args[4])),
					string(tree.MustBeDString(args[5])),
					int(tree.MustBeDInt(args[6])),
					true,
				)
			},
			Info: infoBuilder{info: "Adds a new geometry column to the database, returning metadata about the column created."}.String(),
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"catalog_name", types.String},
				{"schema_name", types.String},
				{"table_name", types.String},
				{"column_name", types.String},
				{"srid", types.Int},
				{"shape", types.String},
				{"num_dimensions", types.Int},
				{"use_typmod", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				return addGeometryColumn(
					ctx,
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					string(tree.MustBeDString(args[2])),
					string(tree.MustBeDString(args[3])),
					int(tree.MustBeDInt(args[4])),
					string(tree.MustBeDString(args[5])),
					int(tree.MustBeDInt(args[6])),
					bool(tree.MustBeDBool(args[7])),
				)
			},
			Info: infoBuilder{info: "Adds a new geometry column to the database, returning metadata about the column created."}.String(),
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"schema_name", types.String},
				{"table_name", types.String},
				{"column_name", types.String},
				{"srid", types.Int},
				{"shape", types.String},
				{"num_dimensions", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				return addGeometryColumn(
					ctx,
					"",
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					string(tree.MustBeDString(args[2])),
					int(tree.MustBeDInt(args[3])),
					string(tree.MustBeDString(args[4])),
					int(tree.MustBeDInt(args[5])),
					true,
				)
			},
			Info: infoBuilder{info: "Adds a new geometry column to the database, returning metadata about the column created."}.String(),
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"schema_name", types.String},
				{"table_name", types.String},
				{"column_name", types.String},
				{"srid", types.Int},
				{"shape", types.String},
				{"num_dimensions", types.Int},
				{"use_typmod", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				return addGeometryColumn(
					ctx,
					"",
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					string(tree.MustBeDString(args[2])),
					int(tree.MustBeDInt(args[3])),
					string(tree.MustBeDString(args[4])),
					int(tree.MustBeDInt(args[5])),
					bool(tree.MustBeDBool(args[6])),
				)
			},
			Info: infoBuilder{info: "Adds a new geometry column to the database, returning metadata about the column created."}.String(),
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"table_name", types.String},
				{"column_name", types.String},
				{"srid", types.Int},
				{"shape", types.String},
				{"num_dimensions", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				return addGeometryColumn(
					ctx,
					"",
					"",
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					int(tree.MustBeDInt(args[2])),
					string(tree.MustBeDString(args[3])),
					int(tree.MustBeDInt(args[4])),
					true,
				)
			},
			Info: infoBuilder{info: "Adds a new geometry column to the database, returning metadata about the column created."}.String(),
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"table_name", types.String},
				{"column_name", types.String},
				{"srid", types.Int},
				{"shape", types.String},
				{"num_dimensions", types.Int},
				{"use_typmod", types.Bool},
			},
			ReturnType: tree.FixedReturnType(types.String),
			Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				return addGeometryColumn(
					ctx,
					"",
					"",
					string(tree.MustBeDString(args[0])),
					string(tree.MustBeDString(args[1])),
					int(tree.MustBeDInt(args[2])),
					string(tree.MustBeDString(args[3])),
					int(tree.MustBeDInt(args[4])),
					bool(tree.MustBeDBool(args[5])),
				)
			},
			Info: infoBuilder{info: "Adds a new geometry column to the database, returning metadata about the column created."}.String(),
		},
	),
}

// geometryOverload2 hides the boilerplate for builtins operating on two geometries.
func geometryOverload2(
	f func(*tree.EvalContext, *tree.DGeometry, *tree.DGeometry) (tree.Datum, error),
	returnType *types.T,
	ib infoBuilder,
) tree.Overload {
	return tree.Overload{
		Types: tree.ArgTypes{
			{"geometry_a", types.Geometry},
			{"geometry_b", types.Geometry},
		},
		ReturnType: tree.FixedReturnType(returnType),
		Fn: func(ctx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
			a := args[0].(*tree.DGeometry)
			b := args[1].(*tree.DGeometry)
			return f(ctx, a, b)
		},
		Info: ib.String(),
	}
}

// geometryOverload2 hides the boilerplate for builtins operating on two geometries
// and the overlap wraps a binary predicate.
func geometryOverload2BinaryPredicate(
	f func(*geo.Geometry, *geo.Geometry) (bool, error), ib infoBuilder,
) tree.Overload {
	return geometryOverload2(
		func(_ *tree.EvalContext, a *tree.DGeometry, b *tree.DGeometry) (tree.Datum, error) {
			ret, err := f(a.Geometry, b.Geometry)
			if err != nil {
				return nil, err
			}
			return tree.MakeDBool(tree.DBool(ret)), nil
		},
		types.Bool,
		ib,
	)
}

// addGeometryColumn appends a geometry column to a table.
func addGeometryColumn(
	ctx *tree.EvalContext,
	catalogName string,
	schemaName string,
	tableName string,
	columnName string,
	srid int,
	shape string,
	dimension int,
	useTypmod bool,
) (tree.Datum, error) {
	if dimension != 2 {
		return nil, pgerror.Newf(
			pgcode.FeatureNotSupported,
			"only dimension=2 is currently supported",
		)
	}
	if !useTypmod {
		return nil, pgerror.Newf(
			pgcode.FeatureNotSupported,
			"useTypmod=false is currently not supported",
		)
	}

	var un tree.UnresolvedName
	if catalogName != "" {
		un = tree.MakeUnresolvedName(catalogName, schemaName, tableName)
	} else if schemaName != "" {
		un = tree.MakeUnresolvedName(schemaName, tableName)
	} else {
		un = tree.MakeUnresolvedName(tableName)
	}

	stmt := `ALTER TABLE %s ADD COLUMN %s GEOMETRY(%s,%d)`
	args := []interface{}{
		un.String(),
		columnName,
		shape,
		srid,
	}

	_, err := ctx.InternalExecutor.Query(
		ctx.Ctx(),
		"addgeometrycolumn",
		ctx.Txn,
		fmt.Sprintf(stmt, args...),
	)
	if err != nil {
		return nil, err
	}
	return tree.NewDString(
		fmt.Sprintf("%s.%s SRID:%d TYPE:%s DIMS:%d", un.String(), columnName, srid, strings.ToUpper(shape), dimension),
	), nil
}

func initGeoBuiltins() {
	for k, v := range geoBuiltins {
		if _, exists := builtins[k]; exists {
			panic("duplicate builtin: " + k)
		}
		v.props.Category = categoryGeospatial
		builtins[k] = v
	}
}
