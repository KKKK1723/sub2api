package service

import (
	"context"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var ErrChannelMonitorOrderConflict = infraerrors.Conflict("CHANNEL_MONITOR_ORDER_CONFLICT", "监控列表或排序已改变，请重新打开排序窗口后再保存")

func (s *ChannelMonitorService) ListOrder(ctx context.Context) ([]ChannelMonitorOrderItem, error) {
	return s.repo.ListOrder(ctx)
}

func (s *ChannelMonitorService) UpdateOrder(ctx context.Context, ids, expectedIDs []int64) error {
	if len(ids) == 0 || len(ids) != len(expectedIDs) {
		return infraerrors.BadRequest("INVALID_MONITOR_ORDER", "必须提交完整的监控排序")
	}
	seen := make(map[int64]bool, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			return infraerrors.BadRequest("INVALID_MONITOR_ORDER", "监控 ID 必须为正数且不能重复")
		}
		seen[id] = true
	}
	for _, id := range expectedIDs {
		if !seen[id] {
			return infraerrors.BadRequest("INVALID_MONITOR_ORDER", "排序前后的监控必须一致")
		}
		delete(seen, id)
	}
	return s.repo.UpdateOrder(ctx, ids, expectedIDs)
}
