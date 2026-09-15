package repository

import (
	"context"
	"slices"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/channelmonitor"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

func (r *channelMonitorRepository) ListOrder(ctx context.Context) ([]service.ChannelMonitorOrderItem, error) {
	items := make([]service.ChannelMonitorOrderItem, 0)
	err := clientFromContext(ctx, r.client).ChannelMonitor.Query().
		Order(dbent.Asc(channelmonitor.FieldSortOrder), dbent.Asc(channelmonitor.FieldID)).
		Select(channelmonitor.FieldID, channelmonitor.FieldName, channelmonitor.FieldGroupName, channelmonitor.FieldEnabled).
		Scan(ctx, &items)
	return items, err
}

func (r *channelMonitorRepository) UpdateOrder(ctx context.Context, ids, expectedIDs []int64) error {
	if tx := dbent.TxFromContext(ctx); tx != nil {
		return updateChannelMonitorOrder(ctx, tx.Client(), ids, expectedIDs)
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := updateChannelMonitorOrder(ctx, tx.Client(), ids, expectedIDs); err != nil {
		return err
	}
	return tx.Commit()
}

func updateChannelMonitorOrder(ctx context.Context, client *dbent.Client, ids, expectedIDs []int64) error {
	// 排序短事务阻止并发增删和排序，避免覆盖其他管理员的保存或漏掉新记录。
	if _, err := client.ExecContext(ctx, "LOCK TABLE channel_monitors IN SHARE ROW EXCLUSIVE MODE"); err != nil {
		return err
	}
	current, err := client.ChannelMonitor.Query().
		Order(dbent.Asc(channelmonitor.FieldSortOrder), dbent.Asc(channelmonitor.FieldID)).IDs(ctx)
	if err != nil {
		return err
	}
	if !slices.Equal(current, expectedIDs) {
		return service.ErrChannelMonitorOrderConflict
	}
	// 序号压缩到 1..N；数据库序列始终领先于记录总数，新建和复制仍会追加末尾。
	_, err = client.ExecContext(ctx, `
		UPDATE channel_monitors AS m SET sort_order = ordered.position
		FROM unnest($1::bigint[]) WITH ORDINALITY AS ordered(id, position)
		WHERE m.id = ordered.id`, pq.Array(ids))
	return err
}
