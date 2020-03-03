// Copyright 2020 The Cockroach Authors.
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
	gosql "database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/internal/client"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlbase"
	"github.com/cockroachdb/cockroach/pkg/sql/tests"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanupSchemaObjects(t *testing.T) {
	defer leaktest.AfterTest(t)()

	ctx := context.Background()
	params, _ := tests.CreateTestServerParams()
	s, db, kvDB := serverutils.StartServer(t, params)
	defer s.Stopper().Stop(ctx)

	conn, err := db.Conn(ctx)
	require.NoError(t, err)

	_, err = conn.ExecContext(ctx, `
SET experimental_enable_temp_tables=true;
SET experimental_serial_normalization='sql_sequence';
CREATE TEMP TABLE a (a SERIAL, c INT);
ALTER TABLE a ADD COLUMN b SERIAL;
CREATE TEMP SEQUENCE a_sequence;
CREATE TEMP VIEW a_view AS SELECT a FROM a;
CREATE TABLE perm_table (a int DEFAULT nextval('a_sequence'), b int);
INSERT INTO perm_table VALUES (DEFAULT, 1);
`)
	require.NoError(t, err)

	rows, err := conn.QueryContext(ctx, `SELECT id, name FROM system.namespace`)
	require.NoError(t, err)

	namesToID := make(map[string]sqlbase.ID)
	var schemaName string
	for rows.Next() {
		var id int64
		var name string
		err := rows.Scan(&id, &name)
		require.NoError(t, err)

		namesToID[name] = sqlbase.ID(id)
		if strings.HasPrefix(name, sessiondata.PgTempSchemaName) {
			schemaName = name
		}
	}

	require.NotEqual(t, "", schemaName)

	tempNames := []string{
		"a",
		"a_view",
		"a_sequence",
		"a_a_seq",
		"a_b_seq",
	}
	selectableTempNames := []string{"a", "a_view"}
	for _, name := range append(tempNames, schemaName) {
		require.Contains(t, namesToID, name)
	}
	for _, name := range selectableTempNames {
		// Check tables are accessible.
		_, err = conn.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %s.%s", schemaName, name))
		require.NoError(t, err)
	}

	require.NoError(
		t,
		kvDB.Txn(ctx, func(ctx context.Context, txn *client.Txn) error {
			err = cleanupSchemaObjects(
				ctx,
				s.ExecutorConfig().(ExecutorConfig).Settings,
				s.InternalExecutor().(*InternalExecutor),
				txn,
				namesToID["defaultdb"],
				schemaName,
			)
			require.NoError(t, err)
			return nil
		}),
	)

	for _, name := range selectableTempNames {
		// Ensure all the entries for the given temporary structures are gone.
		_, err = conn.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %s.%s", schemaName, name))
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf(`relation "%s.%s" does not exist`, schemaName, name))
	}

	// Check perm_table performs correctly, and has the right schema.
	_, err = db.Query("SELECT * FROM perm_table")
	require.NoError(t, err)

	var colDefault gosql.NullString
	err = db.QueryRow(
		`SELECT column_default FROM information_schema.columns
		WHERE table_name = 'perm_table' and column_name = 'a'`,
	).Scan(&colDefault)
	require.NoError(t, err)
	assert.False(t, colDefault.Valid)
}

func TestTemporaryObjectCleanerRetry(t *testing.T) {
	defer leaktest.AfterTest(t)()

	sec := time.Second
	testCases := []struct {
		errTimes       int
		expectedSleeps []time.Duration
		expectedErr    bool
	}{
		{0, []time.Duration{}, false},
		{1, []time.Duration{sec}, false},
		{2, []time.Duration{sec, sec * 2}, false},
		{3, []time.Duration{sec, sec * 2, sec * 4}, false},
		{4, []time.Duration{sec, sec * 2, sec * 4, sec * 8}, false},
		{5, []time.Duration{sec, sec * 2, sec * 4, sec * 8}, true},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("errTimes: %d", tc.errTimes), func(t *testing.T) {
			sleeps := []time.Duration{}
			cleaner := &TemporaryObjectCleaner{
				sleepFunc: func(d time.Duration) {
					sleeps = append(sleeps, d)
				},
			}
			i := 0
			err := cleaner.retry(
				context.Background(),
				func() error {
					if i == tc.errTimes {
						return nil
					}
					i++
					return fmt.Errorf("wooper")
				},
			)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expectedSleeps, sleeps)
		})
	}
}

func TestTemporaryObjectCleaner(t *testing.T) {
	defer leaktest.AfterTest(t)()

	ch := make(chan time.Time)
	knobs := base.TestingKnobs{
		SQLExecutor: &ExecutorTestingKnobs{
			DisableTempObjectsCleanupOnSessionExit: true,
			TempObjectsCleanupCh:                   ch,
		},
	}
	tc := serverutils.StartTestCluster(t, 2, /* numNodes */
		base.TestClusterArgs{
			ServerArgs: base.TestServerArgs{
				UseDatabase: "defaultdb",
				Knobs:       knobs,
			},
		})
	defer tc.Stopper().Stop(context.TODO())

	db := tc.ServerConn(0)
	sqlDB := sqlutils.MakeSQLRunner(db)
	sqlDB.Exec(t, `SET experimental_enable_temp_tables=true`)
	sqlDB.Exec(t, `CREATE TEMP TABLE t (x INT)`)

	// Close the client connection. Normally the temporary data would immediately
	// be cleaned up on session exit, but this is disabled via the
	// DisableTempObjectsCleanupOnSessionExit testing knob.
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	// Sanity check: there should still be one temporary schema present.
	db = tc.ServerConn(1)
	sqlDB = sqlutils.MakeSQLRunner(db)
	tempSchemaQuery := `SELECT count(*) FROM system.namespace WHERE name LIKE 'pg_temp%'`
	var tempSchemaCount int
	sqlDB.QueryRow(t, tempSchemaQuery).Scan(&tempSchemaCount)
	require.Equal(t, tempSchemaCount, 1)

	// Now force a cleanup run (by default, it is every 30mins).
	ch <- timeutil.Now()

	// Verify that the asynchronous cleanup job kicks in and removes the temporary
	// data.
	testutils.SucceedsSoon(t, func() error {
		sqlDB.QueryRow(t, tempSchemaQuery).Scan(&tempSchemaCount)
		if tempSchemaCount > 0 {
			return errors.Errorf("expected 0 temp schemas, found %d", tempSchemaCount)
		}
		return nil
	})
}
