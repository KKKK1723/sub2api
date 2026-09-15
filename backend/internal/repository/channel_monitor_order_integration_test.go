//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

func TestChannelMonitorOrderPersistence(t *testing.T) {
	tx := testEntTx(t)
	ctx := dbent.NewTxContext(context.Background(), tx)
	client := tx.Client()
	_, err := client.ChannelMonitor.Delete().Exec(ctx)
	require.NoError(t, err)
	repo := NewChannelMonitorRepository(client, integrationDB)
	create := func(name string) *service.ChannelMonitor {
		m := &service.ChannelMonitor{Name: name, GroupName: name, Provider: "openai", Endpoint: "https://example.com", APIKey: "encrypted", PrimaryModel: "gpt-5", Enabled: true, IntervalSeconds: 60, CreatedBy: 1}
		require.NoError(t, repo.Create(ctx, m))
		return m
	}
	a, b := create("Plus"), create("Pro")
	orderIDs := func() []int64 {
		items, err := repo.ListOrder(ctx)
		require.NoError(t, err)
		ids := make([]int64, 0, len(items))
		for _, item := range items {
			ids = append(ids, item.ID)
		}
		return ids
	}
	require.Equal(t, []int64{a.ID, b.ID}, orderIDs())
	require.NoError(t, repo.UpdateOrder(ctx, []int64{b.ID, a.ID}, []int64{a.ID, b.ID}))
	// 普通配置更新即使携带旧排序字段，也不能覆盖管理员已保存的顺序。
	a.Name, a.GroupName, a.Enabled = "renamed", "Kiro", false
	require.NoError(t, repo.Update(ctx, a))
	require.NoError(t, repo.MarkChecked(ctx, b.ID, time.Now()))
	require.Equal(t, []int64{b.ID, a.ID}, orderIDs())
	enabled, err := repo.ListEnabled(ctx)
	require.NoError(t, err)
	require.Len(t, enabled, 1)
	require.Equal(t, b.ID, enabled[0].ID)
	a.Enabled = true
	require.NoError(t, repo.Update(ctx, a))
	c := create("OpenAI 官 key1")
	require.Equal(t, []int64{b.ID, a.ID, c.ID}, orderIDs())
	// 过期窗口不能覆盖刚加入的记录或另一位管理员的顺序。
	require.ErrorIs(t, repo.UpdateOrder(ctx, []int64{a.ID, b.ID}, []int64{b.ID, a.ID}), service.ErrChannelMonitorOrderConflict)
	require.Equal(t, []int64{b.ID, a.ID, c.ID}, orderIDs())
	enabled, err = repo.ListEnabled(ctx)
	require.NoError(t, err)
	require.Equal(t, []int64{b.ID, a.ID, c.ID}, []int64{enabled[0].ID, enabled[1].ID, enabled[2].ID})
	page, _, err := repo.List(ctx, service.ChannelMonitorListParams{Page: 1, PageSize: 2})
	require.NoError(t, err)
	require.Equal(t, []int64{b.ID, a.ID}, []int64{page[0].ID, page[1].ID})
}

func TestChannelMonitorOrderMigration(t *testing.T) {
	tx := testEntTx(t)
	ctx := context.Background()
	client := tx.Client()
	_, err := client.ExecContext(ctx, `CREATE SCHEMA monitor_order_migration_test; SET LOCAL search_path TO monitor_order_migration_test;
		CREATE TABLE channel_monitors (id BIGSERIAL PRIMARY KEY, name TEXT NOT NULL, group_name TEXT);
		INSERT INTO channel_monitors (name, group_name) VALUES
		('Claude-kiro反代', NULL), ('GPT-Pro混池', ''), ('GPT-Plus', ''), ('GPT-Pro', ''), ('GPT-Pro满血', ''),
		('monitor2', 'OpenAI官key2'), ('monitor1', 'OpenAI官key1');`)
	require.NoError(t, err)
	query, err := migrations.FS.ReadFile("193_channel_monitor_sort_order.sql")
	require.NoError(t, err)
	_, err = client.ExecContext(ctx, string(query))
	require.NoError(t, err)
	readIDs := func() []int64 {
		rows, err := client.QueryContext(ctx, "SELECT id FROM channel_monitors ORDER BY sort_order, id")
		require.NoError(t, err)
		defer rows.Close()
		var ids []int64
		for rows.Next() {
			var id int64
			require.NoError(t, rows.Scan(&id))
			ids = append(ids, id)
		}
		require.NoError(t, rows.Err())
		return ids
	}
	require.Equal(t, []int64{7, 6, 5, 4, 2, 3, 1}, readIDs())
	_, err = client.ExecContext(ctx, "UPDATE channel_monitors SET sort_order = 8 - sort_order")
	require.NoError(t, err)
	_, err = client.ExecContext(ctx, string(query))
	require.NoError(t, err)
	require.Equal(t, []int64{1, 3, 2, 4, 5, 6, 7}, readIDs())
	_, err = client.ExecContext(ctx, "INSERT INTO channel_monitors (name) VALUES ('OpenAI官key1')")
	require.NoError(t, err)
	require.Equal(t, []int64{1, 3, 2, 4, 5, 6, 7, 8}, readIDs())
}
