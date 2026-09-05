package repository

import (
	"context"
	"regexp"
	"testing"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUpstreamBalanceExtraIsSchedulerNeutral(t *testing.T) {
	require.True(t, isSchedulerNeutralExtraKey(service.UpstreamBalanceProbeEnabledExtraKey))
	require.True(t, isSchedulerNeutralExtraKey(service.UpstreamBalanceQueryExtraKey))
	require.True(t, isSchedulerNeutralExtraKey(service.UpstreamBalanceProbeExtraKey))
}

func TestLockAndMergeAccountProbeExtraProtectsCurrentUpstreamBalanceSnapshot(t *testing.T) {
	for _, tt := range []struct {
		name         string
		requested    any
		wantSnapshot bool
	}{
		{name: "same query preserves the latest database snapshot", requested: map[string]any{"preset": "ccswitch"}, wantSnapshot: true},
		{name: "changed query clears the previous snapshot", requested: map[string]any{"preset": "custom"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			t.Cleanup(func() { _ = db.Close() })
			client := dbent.NewClient(dbent.Driver(sql.OpenDB(dialect.Postgres, db)))
			t.Cleanup(func() { _ = client.Close() })

			mock.ExpectQuery(`(?s)`+regexp.QuoteMeta("SELECT")+`.*`+regexp.QuoteMeta("FOR NO KEY UPDATE")).
				WithArgs(int64(31), service.PlatformAnthropic, service.AccountTypeAPIKey, `{"api_key":"sk-test"}`, nil).
				WillReturnRows(sqlmock.NewRows([]string{
					"identity_unchanged", "ollama_group_unchanged", "ollama_proxy_unchanged",
					"enabled", "snapshot", "balance_query", "balance_snapshot",
					"ollama_session", "ollama_auto", "ollama_snapshot",
				}).AddRow(
					true, false, true, nil, nil,
					[]byte(`{"preset":"ccswitch"}`), []byte(`{"status":"ok","data":{"balance":12.5,"unit":"USD"}}`),
					nil, nil, nil,
				))

			account := &service.Account{
				ID: 31, Platform: service.PlatformAnthropic, Type: service.AccountTypeAPIKey,
				Credentials: map[string]any{"api_key": "sk-test"},
				Extra: map[string]any{
					service.UpstreamBalanceQueryExtraKey: tt.requested,
					service.UpstreamBalanceProbeExtraKey: map[string]any{"status": "forged"},
				},
			}
			got, err := lockAndMergeAccountProbeExtra(context.Background(), client, account, nil)
			require.NoError(t, err)
			if tt.wantSnapshot {
				require.Equal(t, map[string]any{
					"status": "ok",
					"data":   map[string]any{"balance": 12.5, "unit": "USD"},
				}, got[service.UpstreamBalanceProbeExtraKey])
			} else {
				require.NotContains(t, got, service.UpstreamBalanceProbeExtraKey)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateCredentialsClearsUpstreamBalanceSnapshotForAPIKeyIdentityChange(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	client := dbent.NewClient(dbent.Driver(sql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })

	mock.ExpectBegin()
	mock.ExpectExec(`(?s)UPDATE accounts.*type = 'apikey'.*credentials IS DISTINCT FROM \$1::jsonb.*- 'upstream_balance_probe'`).
		WithArgs(`{"api_key":"changed"}`, int64(27)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO scheduler_outbox")).
		WithArgs(service.SchedulerOutboxEventAccountChanged, int64(27), nil, nil, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := newAccountRepositoryWithSQL(client, db, nil)
	require.NoError(t, repo.UpdateCredentials(context.Background(), 27, map[string]any{"api_key": "changed"}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBulkProxyUpdateClearsUpstreamBalanceSnapshotForAPIKeys(t *testing.T) {
	exec := &recordingSQLExecutor{result: rowsAffectedResult(1)}
	repo := newAccountRepositoryWithSQL(nil, exec, nil)
	proxyID := int64(9)

	_, err := repo.BulkUpdate(context.Background(), []int64{27}, service.AccountBulkUpdate{ProxyID: &proxyID})

	require.NoError(t, err)
	require.NotEmpty(t, exec.execQueries)
	query := normalizeSQLWhitespace(exec.execQueries[0])
	require.Contains(t, query, "type = 'apikey' AND proxy_id IS DISTINCT FROM $1")
	require.Contains(t, query, "- 'upstream_balance_probe'")
}

func TestBulkCredentialUpdateClearsUpstreamBalanceSnapshotForAPIKeys(t *testing.T) {
	exec := &recordingSQLExecutor{result: rowsAffectedResult(1)}
	repo := newAccountRepositoryWithSQL(nil, exec, nil)

	_, err := repo.BulkUpdate(context.Background(), []int64{27}, service.AccountBulkUpdate{
		Credentials: map[string]any{"base_url": "https://upstream.example/v1"},
	})

	require.NoError(t, err)
	require.NotEmpty(t, exec.execQueries)
	query := normalizeSQLWhitespace(exec.execQueries[0])
	require.Contains(t, query, "type = 'apikey' AND credentials IS DISTINCT FROM")
	require.Contains(t, query, "- 'upstream_balance_probe'")
}
