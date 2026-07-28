package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func init() {
	schedulerNeutralExtraKeyPrefixes = append(schedulerNeutralExtraKeyPrefixes, "upstream_balance_")
}

func (r *accountRepository) UpdateUpstreamBalanceProbeSnapshot(
	ctx context.Context,
	account *service.Account,
	snapshot *service.UpstreamBalanceProbeSnapshot,
) error {
	if account == nil || snapshot == nil {
		return service.ErrAccountNilInput
	}
	if dbent.TxFromContext(ctx) == nil {
		tx, err := r.client.Tx(ctx)
		if errors.Is(err, dbent.ErrTxStarted) {
			return r.updateUpstreamBalanceProbeSnapshotInTx(ctx, account, snapshot)
		}
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()
		if err := r.updateUpstreamBalanceProbeSnapshotInTx(dbent.NewTxContext(ctx, tx), account, snapshot); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		r.syncSchedulerAccountSnapshot(ctx, account.ID)
		return nil
	}
	return r.updateUpstreamBalanceProbeSnapshotInTx(ctx, account, snapshot)
}

func (r *accountRepository) updateUpstreamBalanceProbeSnapshotInTx(
	ctx context.Context,
	account *service.Account,
	snapshot *service.UpstreamBalanceProbeSnapshot,
) error {
	payload, err := json.Marshal(map[string]any{service.UpstreamBalanceProbeExtraKey: snapshot})
	if err != nil {
		return err
	}
	credentials, err := json.Marshal(account.Credentials)
	if err != nil {
		return err
	}
	expected := func(key string) (string, error) {
		var value any
		if account.Extra != nil {
			value = account.Extra[key]
		}
		raw, marshalErr := json.Marshal(value)
		return string(raw), marshalErr
	}
	expectedSnapshot, err := expected(service.UpstreamBalanceProbeExtraKey)
	if err != nil {
		return err
	}
	expectedEnabled, err := expected(service.UpstreamBalanceProbeEnabledExtraKey)
	if err != nil {
		return err
	}
	expectedQuery, err := expected(service.UpstreamBalanceQueryExtraKey)
	if err != nil {
		return err
	}

	client := clientFromContext(ctx, r.client)
	proxyMatches, err := lockAndMatchProbeProxyIdentity(ctx, client, account)
	if err != nil {
		return err
	}
	if !proxyMatches {
		return service.ErrUpstreamBalanceProbeIdentityChanged
	}
	var proxyID any
	if account.ProxyID != nil {
		proxyID = *account.ProxyID
	}
	result, err := client.ExecContext(ctx, `
		UPDATE accounts
		SET extra = COALESCE(extra, '{}'::jsonb) || $1::jsonb, updated_at = NOW()
		WHERE id = $2
			AND platform = $3
			AND type = $4
			AND credentials = $5::jsonb
			AND proxy_id IS NOT DISTINCT FROM $6
			AND COALESCE(extra -> 'upstream_balance_probe', 'null'::jsonb) = $7::jsonb
			AND COALESCE(extra -> 'upstream_balance_probe_enabled', 'null'::jsonb) = $8::jsonb
			AND COALESCE(extra -> 'upstream_balance_query', 'null'::jsonb) = $9::jsonb
			AND deleted_at IS NULL
	`, string(payload), account.ID, account.Platform, account.Type, string(credentials), proxyID, expectedSnapshot, expectedEnabled, expectedQuery)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrUpstreamBalanceProbeIdentityChanged
	}
	return enqueueSchedulerOutbox(ctx, client, service.SchedulerOutboxEventAccountChanged, &account.ID, nil, nil)
}

func (r *accountRepository) ListDueUpstreamBalanceProbeAccounts(ctx context.Context, now time.Time, limit int) ([]service.Account, error) {
	if limit <= 0 {
		return []service.Account{}, nil
	}
	if r.sql == nil {
		return nil, errors.New("account repository SQL executor not configured")
	}
	rows, err := r.sql.QueryContext(ctx, `
		WITH candidates AS (
			SELECT
				id,
				extra #>> '{upstream_balance_probe,status}' AS probe_status,
				extra #>> '{upstream_balance_probe,next_probe_at}' AS next_probe_at
			FROM accounts
			WHERE deleted_at IS NULL
				AND status = 'active'
				AND type = 'apikey'
				AND extra @> '{"upstream_balance_probe_enabled": true}'::jsonb
				AND extra -> 'upstream_balance_query' IS NOT NULL
		), parsed AS MATERIALIZED (
			SELECT
				id,
				probe_status,
				next_probe_at,
				next_probe_at ~ '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]+)?(Z|[+-][0-9]{2}:[0-9]{2})$' AS rfc3339_shape,
				jsonb_path_query_first_tz(
					jsonb_build_object(
						'value',
						replace(regexp_replace(regexp_replace(
							next_probe_at,
							'(\.[0-9]{6})[0-9]+(Z|[+-][0-9]{2}:[0-9]{2})$',
							'\1\2'
						), 'Z$', '+00:00'), 'T', ' ')
					),
					'$.value.datetime()',
					'{}'::jsonb,
					true
				) #>> '{}' AS parsed_next_probe_at
			FROM candidates
		), normalized AS (
			SELECT
				id,
				probe_status,
				next_probe_at,
				parsed_next_probe_at,
				rfc3339_shape AND parsed_next_probe_at IS NOT NULL AS valid_next_probe_at
			FROM parsed
		)
		SELECT id
		FROM normalized
		WHERE probe_status NOT IN ('ok', 'unsupported', 'failed')
			OR probe_status IS NULL
			OR next_probe_at IS NULL
			OR NOT valid_next_probe_at
			OR CASE WHEN valid_next_probe_at THEN parsed_next_probe_at::timestamptz <= $1 ELSE FALSE END
		ORDER BY
			CASE
				WHEN probe_status NOT IN ('ok', 'unsupported', 'failed')
					OR probe_status IS NULL
					OR next_probe_at IS NULL
					OR NOT valid_next_probe_at
				THEN 0
				ELSE 1
			END ASC,
			CASE WHEN valid_next_probe_at THEN parsed_next_probe_at::timestamptz END ASC NULLS FIRST,
			id ASC
		LIMIT $2
	`, now.UTC(), limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	ids := make([]int64, 0, limit)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []service.Account{}, nil
	}
	accounts, err := r.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]service.Account, 0, len(accounts))
	for _, account := range accounts {
		if account != nil {
			out = append(out, *account)
		}
	}
	return out, nil
}
