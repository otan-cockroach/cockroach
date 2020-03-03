// Copyright 2019 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package sql

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/internal/client"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/schema"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlbase"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/uint128"
	"github.com/cockroachdb/errors"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

// defaultTempObjectCleanupInterval is the interval of time after which a background job
// runs to cleanup temporary tables that weren't deleted at session exit.
var defaultTempObjectCleanupInterval = 30 * time.Minute

func createTempSchema(params runParams, sKey sqlbase.DescriptorKey) (sqlbase.ID, error) {
	id, err := GenerateUniqueDescID(params.ctx, params.extendedEvalCtx.ExecCfg.DB)
	if err != nil {
		return sqlbase.InvalidID, err
	}
	if err := params.p.createSchemaWithID(params.ctx, sKey.Key(), id); err != nil {
		return sqlbase.InvalidID, err
	}

	params.p.sessionDataMutator.SetTemporarySchemaName(sKey.Name())
	return id, nil
}

func (p *planner) createSchemaWithID(
	ctx context.Context, schemaNameKey roachpb.Key, schemaID sqlbase.ID,
) error {
	if p.ExtendedEvalContext().Tracing.KVTracingEnabled() {
		log.VEventf(ctx, 2, "CPut %s -> %d", schemaNameKey, schemaID)
	}

	b := &client.Batch{}
	b.CPut(schemaNameKey, schemaID, nil)

	return p.txn.Run(ctx, b)
}

// temporarySchemaName returns the session specific temporary schema name given
// the sessionID. When the session creates a temporary object for the first
// time, it must create a schema with the name returned by this function.
func temporarySchemaName(sessionID ClusterWideID) string {
	return fmt.Sprintf("pg_temp_%d_%d", sessionID.Hi, sessionID.Lo)
}

// temporarySchemaSessionID returns the sessionID of the given temporary schema.
func temporarySchemaSessionID(scName string) (bool, ClusterWideID, error) {
	if !strings.HasPrefix(scName, "pg_temp_") {
		return false, ClusterWideID{}, nil
	}
	parts := strings.Split(scName, "_")
	if len(parts) != 4 {
		return false, ClusterWideID{}, errors.Errorf("malformed temp schema name %s", scName)
	}
	hi, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return false, ClusterWideID{}, err
	}
	lo, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		return false, ClusterWideID{}, err
	}
	return true, ClusterWideID{uint128.Uint128{Hi: hi, Lo: lo}}, nil
}

// getTemporaryObjectNames returns all the temporary objects under the
// temporary schema of the given dbID.
func getTemporaryObjectNames(
	ctx context.Context, txn *client.Txn, dbID sqlbase.ID, tempSchemaName string,
) (TableNames, error) {
	dbDesc, err := MustGetDatabaseDescByID(ctx, txn, dbID)
	if err != nil {
		return nil, err
	}
	a := UncachedPhysicalAccessor{}
	return a.GetObjectNames(
		ctx,
		txn,
		dbDesc,
		tempSchemaName,
		tree.DatabaseListFlags{CommonLookupFlags: tree.CommonLookupFlags{Required: false}},
	)
}

// cleanupSessionTempObjects removes all temporary objects (tables, sequences,
// views, temporary schema) created by the session.
func cleanupSessionTempObjects(
	ctx context.Context,
	settings *cluster.Settings,
	db *client.DB,
	ie sqlutil.InternalExecutor,
	sessionID ClusterWideID,
) error {
	tempSchemaName := temporarySchemaName(sessionID)
	return db.Txn(ctx, func(ctx context.Context, txn *client.Txn) error {
		// We are going to read all database descriptor IDs, then for each database
		// we will drop all the objects under the temporary schema.
		dbIDs, err := GetAllDatabaseDescriptorIDs(ctx, txn)
		if err != nil {
			return err
		}
		for _, id := range dbIDs {
			if err := cleanupSchemaObjects(
				ctx,
				settings,
				ie,
				txn,
				id,
				tempSchemaName,
			); err != nil {
				return err
			}
			// Even if no objects were found under the temporary schema, the schema
			// itself may still exist (eg. a temporary table was created and then
			// dropped). So we remove the namespace table entry of the temporary
			// schema.
			if err := sqlbase.RemoveSchemaNamespaceEntry(ctx, txn, id, tempSchemaName); err != nil {
				return err
			}
		}
		return nil
	})
}

// cleanupSchemaObjects removes all objects that is located within a dbID and schema.
func cleanupSchemaObjects(
	ctx context.Context,
	settings *cluster.Settings,
	ie sqlutil.InternalExecutor,
	txn *client.Txn,
	dbID sqlbase.ID,
	schemaName string,
) error {
	tbNames, err := getTemporaryObjectNames(ctx, txn, dbID, schemaName)
	if err != nil {
		return err
	}
	a := UncachedPhysicalAccessor{}

	searchPath := sqlbase.DefaultSearchPath.WithTemporarySchemaName(schemaName)
	override := sqlbase.InternalExecutorSessionDataOverride{
		SearchPath: &searchPath,
		User:       security.RootUser,
	}

	// TODO(andrei): We might want to accelerate the deletion of this data.
	var tables sqlbase.IDs
	var views sqlbase.IDs
	var sequences sqlbase.IDs

	descsByID := make(map[sqlbase.ID]*TableDescriptor, len(tbNames))
	tblNamesByID := make(map[sqlbase.ID]tree.TableName, len(tbNames))
	for _, tbName := range tbNames {
		objDesc, err := a.GetObjectDesc(
			ctx,
			txn,
			settings,
			&tbName,
			tree.ObjectLookupFlagsWithRequired(),
		)
		if err != nil {
			return err
		}
		desc := objDesc.TableDesc()

		descsByID[desc.ID] = desc
		tblNamesByID[desc.ID] = tbName

		if desc.SequenceOpts != nil {
			sequences = append(sequences, desc.ID)
		} else if desc.ViewQuery != "" {
			views = append(views, desc.ID)
		} else {
			tables = append(tables, desc.ID)
		}
	}

	for _, toDelete := range []struct {
		// typeName is the type of table being deleted, e.g. view, table, sequence
		typeName string
		// ids represents which ids we wish to remove.
		ids sqlbase.IDs
		// preHook is used to perform any operations needed before calling
		// delete on all the given ids.
		preHook func(sqlbase.ID) error
	}{
		// Drop views before tables to avoid deleting required dependencies.
		{"VIEW", views, nil},
		{"TABLE", tables, nil},
		// Drop sequences after tables, because then we reduce the amount of work
		// that may be needed to drop indices.
		{
			"SEQUENCE",
			sequences,
			func(id sqlbase.ID) error {
				desc := descsByID[id]
				// For any dependent tables, we need to drop the sequence dependencies.
				// This can happen if a permanent table references a temporary table.
				for _, d := range desc.DependedOnBy {
					// We have already cleaned out anything we are depended on if we've seen
					// the descriptor already.
					if _, ok := descsByID[d.ID]; ok {
						continue
					}
					dTableDesc, err := sqlbase.GetTableDescFromID(ctx, txn, d.ID)
					if err != nil {
						return err
					}
					db, err := sqlbase.GetDatabaseDescFromID(ctx, txn, dTableDesc.GetParentID())
					if err != nil {
						return err
					}
					schema, err := schema.ResolveNameByID(
						ctx,
						txn,
						dTableDesc.GetParentID(),
						dTableDesc.GetParentSchemaID(),
					)
					if err != nil {
						return err
					}
					dependentColIDs := util.MakeFastIntSet()
					for _, colID := range d.ColumnIDs {
						dependentColIDs.Add(int(colID))
					}
					for _, col := range dTableDesc.Columns {
						if dependentColIDs.Contains(int(col.ID)) {
							tbName := tree.MakeTableNameWithSchema(
								tree.Name(db.Name),
								tree.Name(schema),
								tree.Name(dTableDesc.Name),
							)
							_, err = ie.ExecEx(
								ctx,
								"delete-temp-dependent-col",
								txn,
								override,
								fmt.Sprintf(
									"ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT",
									tbName.FQString(),
									tree.NameString(col.Name),
								),
							)
							if err != nil {
								return err
							}
						}
					}
				}
				return nil
			},
		},
	} {
		if len(toDelete.ids) > 0 {
			if toDelete.preHook != nil {
				for _, id := range toDelete.ids {
					if err := toDelete.preHook(id); err != nil {
						return err
					}
				}
			}

			var query strings.Builder
			query.WriteString("DROP ")
			query.WriteString(toDelete.typeName)

			for i, id := range toDelete.ids {
				tbName := tblNamesByID[id]
				if i != 0 {
					query.WriteString(",")
				}
				query.WriteString(" ")
				query.WriteString(tbName.FQString())
			}
			query.WriteString(" CASCADE")
			_, err = ie.ExecEx(ctx, "delete-temp-"+toDelete.typeName, txn, override, query.String())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// isMeta1LeaseholderFunc helps us avoid an import into pkg/storage.
type isMeta1LeaseholderFunc func(hlc.Timestamp) (bool, error)

// TemporaryObjectCleaner is a background thread job that periodically
// cleans up orphaned temporary objects by sessions which did not close
// down cleanly.
type TemporaryObjectCleaner struct {
	settings                         *cluster.Settings
	db                               *client.DB
	makeSessionBoundInternalExecutor sqlutil.SessionBoundInternalExecutorFactory
	statusServer                     serverpb.StatusServer
	isMeta1LeaseholderFunc           isMeta1LeaseholderFunc
	testingKnobs                     ExecutorTestingKnobs
	metrics                          *temporaryObjectCleanerMetrics

	sleepFunc func(time.Duration)
}

// temporaryObjectCleanerMetrics are the metrics for TemporaryObjectCleaner
type temporaryObjectCleanerMetrics struct {
	jobRunning             *metric.Gauge
	schemasToDelete        *metric.Counter
	schemasDeletionError   *metric.Counter
	schemasDeletionSuccess *metric.Counter
}

// NewTemporaryObjectCleaner initializes the TemporaryObjectCleaner with the
// required arguments, but does not start it.
func NewTemporaryObjectCleaner(
	settings *cluster.Settings,
	db *client.DB,
	makeSessionBoundInternalExecutor sqlutil.SessionBoundInternalExecutorFactory,
	statusServer serverpb.StatusServer,
	isMeta1LeaseholderFunc isMeta1LeaseholderFunc,
	testingKnobs ExecutorTestingKnobs,
) *TemporaryObjectCleaner {
	return &TemporaryObjectCleaner{
		settings:                         settings,
		db:                               db,
		makeSessionBoundInternalExecutor: makeSessionBoundInternalExecutor,
		statusServer:                     statusServer,
		isMeta1LeaseholderFunc:           isMeta1LeaseholderFunc,
		testingKnobs:                     testingKnobs,
		metrics: &temporaryObjectCleanerMetrics{
			jobRunning: metric.NewGauge(metric.Metadata{
				Name:        "sql.temp_object_cleaner.jobs_running",
				Help:        "number of cleaner jobs running on this node",
				Measurement: "Count",
				Unit:        metric.Unit_COUNT,
				MetricType:  io_prometheus_client.MetricType_GAUGE,
			}),
			schemasToDelete: metric.NewCounter(metric.Metadata{
				Name:        "sql.temp_object_cleaner.schemas_to_delete",
				Help:        "number of schemas to be deleted by the temp object cleaner on this node",
				Measurement: "Count",
				Unit:        metric.Unit_COUNT,
				MetricType:  io_prometheus_client.MetricType_COUNTER,
			}),
			schemasDeletionError: metric.NewCounter(metric.Metadata{
				Name:        "sql.temp_object_cleaner.schemas_deletion_error",
				Help:        "number of errored schema deletions by the temp object cleaner on this node",
				Measurement: "Count",
				Unit:        metric.Unit_COUNT,
				MetricType:  io_prometheus_client.MetricType_COUNTER,
			}),
			schemasDeletionSuccess: metric.NewCounter(metric.Metadata{
				Name:        "sql.temp_object_cleaner.schemas_deletion_success",
				Help:        "number of successful schema deletions by the temp object cleaner on this node",
				Measurement: "Count",
				Unit:        metric.Unit_COUNT,
				MetricType:  io_prometheus_client.MetricType_COUNTER,
			}),
		},
		sleepFunc: time.Sleep,
	}
}

// retry wraps a closure allowing retries, with exponential backoff.
func (c *TemporaryObjectCleaner) retry(ctx context.Context, do func() error) error {
	var err error
	retrySec := 1 * time.Second
	for i := 0; i < 5; i++ {
		if i > 0 {
			c.sleepFunc(retrySec)
			retrySec *= 2
		}
		if err = do(); err == nil {
			return err
		}
		log.Warningf(ctx, "temporary schema cleanup: attempt #%d: error: %v", i, err)
	}
	return err
}

// doTemporaryObjectCleanup performs the actual cleanup.
func (c *TemporaryObjectCleaner) doTemporaryObjectCleanup(ctx context.Context) error {

	// We only want to perform the cleanup if we are holding the meta1 lease.
	// This ensures only one server can perform the job at a time.
	var isLeaseholder bool
	if err := c.retry(ctx, func() error {
		var err error
		isLeaseholder, err = c.isMeta1LeaseholderFunc(c.db.Clock().Now())
		return err
	}); err != nil {
		return err
	}
	if !isLeaseholder {
		log.Infof(ctx, "skipping temporary object cleanup run as it is not the leaseholder")
		return nil
	}

	c.metrics.jobRunning.Inc(1)
	defer c.metrics.jobRunning.Dec(1)

	log.Infof(ctx, "running temporary object cleanup background job")
	txn := client.NewTxn(ctx, c.db, 0)

	// Build a set of all session IDs with temporary objects.
	var dbIDs []sqlbase.ID
	if err := c.retry(ctx, func() error {
		var err error
		dbIDs, err = GetAllDatabaseDescriptorIDs(ctx, txn)
		return err
	}); err != nil {
		return err
	}
	sessionIDs := make(map[ClusterWideID]struct{})
	for _, dbID := range dbIDs {
		var schemaNames map[sqlbase.ID]string
		if err := c.retry(ctx, func() error {
			var err error
			schemaNames, err = schema.GetForDatabase(ctx, txn, dbID)
			return err
		}); err != nil {
			return err
		}
		for _, scName := range schemaNames {
			isTempSchema, sessionID, err := temporarySchemaSessionID(scName)
			if err != nil {
				// This should not cause an error.
				log.Warningf(ctx, "could not parse %q as temporary schema name", scName)
				continue
			}
			if isTempSchema {
				sessionIDs[sessionID] = struct{}{}
			}
		}
	}

	// Get active sessions.
	var response *serverpb.ListSessionsResponse
	if err := c.retry(ctx, func() error {
		var err error
		response, err = c.statusServer.ListSessions(
			ctx,
			&serverpb.ListSessionsRequest{},
		)
		return err
	}); err != nil {
		return err
	}
	activeSessions := make(map[uint128.Uint128]struct{})
	for _, session := range response.Sessions {
		activeSessions[uint128.FromBytes(session.ID)] = struct{}{}
	}

	// Clean up temporary data for inactive sessions.
	ie := c.makeSessionBoundInternalExecutor(ctx, &sessiondata.SessionData{})
	for sessionID := range sessionIDs {
		if _, ok := activeSessions[sessionID.Uint128]; !ok {
			log.Infof(ctx, "cleaning up temporary object for session %q", sessionID)
			c.metrics.schemasToDelete.Inc(1)

			// Reset the session data with the appropriate sessionID such that we can resolve
			// the given schema correctly.
			if err := c.retry(ctx, func() error {
				return cleanupSessionTempObjects(
					ctx,
					c.settings,
					c.db,
					ie,
					sessionID,
				)
			}); err != nil {
				// Log error but continue trying to delete the rest.
				log.Warningf(ctx, "failed to clean temp objects under session %q: %v", sessionID, err)
				c.metrics.schemasDeletionError.Inc(1)
			} else {
				c.metrics.schemasDeletionSuccess.Inc(1)
				telemetry.Inc(sqltelemetry.TempObjectCleanerDeletionCounter)
			}
		}
	}

	log.Infof(ctx, "completed temporary object cleanup job")
	return nil
}

// Start initializes the background thread which periodically cleans up leftover temporary objects.
func (c *TemporaryObjectCleaner) Start(ctx context.Context, stopper *stop.Stopper) {
	stopper.RunWorker(ctx, func(ctx context.Context) {
		ticker := time.NewTicker(defaultTempObjectCleanupInterval)
		defer ticker.Stop()
		tickCh := ticker.C
		if c.testingKnobs.TempObjectsCleanupCh != nil {
			tickCh = c.testingKnobs.TempObjectsCleanupCh
		}

		for {
			select {
			case <-tickCh:
				if err := c.doTemporaryObjectCleanup(ctx); err != nil {
					log.Warningf(ctx, "failed to clean temp objects: %v", err)
				}
			case <-stopper.ShouldQuiesce():
				return
			case <-ctx.Done():
				return
			}
		}
	})
}
