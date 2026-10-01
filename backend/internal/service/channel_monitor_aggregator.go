package service

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"
)

// 渠道监控聚合层：把 latest + availability 拼成 admin/user 视图所需的 summary / detail。
// 所有方法都遵守"失败仅日志，返回零值"的原则，避免 N+1 查询失败拖垮列表渲染。

// BatchMonitorStatusSummary 批量聚合多个监控的 latest + 7d 可用率（admin/user list 用，消除 N+1）。
// 失败时返回空 map，错误仅日志，不影响列表渲染。
//
// 参数：
//   - ids: 要聚合的 monitor ID 列表
//   - primaryByID: monitor ID -> primary model（用于读 7d 可用率与 latest 状态）
//   - extrasByID: monitor ID -> extra models 列表（用于读 latest 状态填充 ExtraModels）
func (s *ChannelMonitorService) BatchMonitorStatusSummary(
	ctx context.Context,
	ids []int64,
	primaryByID map[int64]string,
	extrasByID map[int64][]string,
	probeModelsByID map[int64][]string,
) map[int64]MonitorStatusSummary {
	out := make(map[int64]MonitorStatusSummary, len(ids))
	if len(ids) == 0 {
		return out
	}
	latestMap, err := s.repo.ListLatestForMonitorIDs(ctx, ids)
	if err != nil {
		slog.Warn("channel_monitor: batch load latest failed", "error", err)
		latestMap = map[int64][]*ChannelMonitorLatest{}
	}
	availMap, err := s.repo.ComputeAvailabilityForMonitors(ctx, ids, monitorAvailability7Days)
	if err != nil {
		slog.Warn("channel_monitor: batch compute availability failed", "error", err)
		availMap = map[int64][]*ChannelMonitorAvailability{}
	}

	for _, id := range ids {
		out[id] = buildStatusSummary(
			indexLatestByModel(latestMap[id]),
			indexAvailabilityByModel(availMap[id]),
			primaryByID[id],
			extrasByID[id],
			probeModelsByID[id],
			latestMap[id],
		)
	}
	return out
}

// ListUserView 用户只读视图：列出所有 enabled 监控的概览。
// 使用批量聚合接口避免 N+1：
//
//	1 次查 monitors；
//	1 次批量 latest（含 ping_latency_ms）；
//	1 次批量 7d availability；
//	1 次批量 timeline（主模型最近 N 条）。
func (s *ChannelMonitorService) ListUserView(ctx context.Context) ([]*UserMonitorView, error) {
	monitors, err := s.repo.ListEnabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("list enabled monitors: %w", err)
	}
	if len(monitors) == 0 {
		return []*UserMonitorView{}, nil
	}

	ids, primaryByID, extrasByID, probeModelsByID := collectMonitorIndexes(monitors)
	summaries := s.BatchMonitorStatusSummary(ctx, ids, primaryByID, extrasByID, probeModelsByID)
	latestMap := s.batchLatest(ctx, ids)
	timelineMap := s.batchTimeline(ctx, ids, primaryByID)

	views := make([]*UserMonitorView, 0, len(monitors))
	for _, m := range monitors {
		if probeStatus := MonitorConfiguredProbeStatus(m); probeStatus != nil {
			summary := summaries[m.ID]
			summary.PrimaryStatus = probeStatus.Status
			summary.PrimaryLatencyMs = probeStatus.LatencyMs
			summaries[m.ID] = summary
		}
		primaryLatest := pickLatest(latestMap[m.ID], m.PrimaryModel)
		if aggregateLatest := pickLatest(latestMap[m.ID], MonitorAggregateModel); aggregateLatest != nil {
			primaryLatest = aggregateLatest
		}
		views = append(views, buildUserViewFromSummary(m, summaries[m.ID], primaryLatest, timelineMap[m.ID]))
	}
	return views, nil
}

// MonitorConfiguredProbeStatus 返回监控已保存探针状态的聚合结果。
// quota 模式不做探活，保留的探针配置只用于切回探活模式，其中的状态已过期，不参与展示。
func MonitorConfiguredProbeStatus(m *ChannelMonitor) *CheckResult {
	if m == nil || m.CheckMode == MonitorCheckModeQuota {
		return nil
	}
	return AggregateConfiguredProbeStatuses(m.Probes)
}

// AggregateConfiguredProbeStatuses 按绿、黄、红优先级聚合当前已保存的探针状态。
func AggregateConfiguredProbeStatuses(probes []MonitorProbe) *CheckResult {
	var best *CheckResult
	rank := func(status string) int {
		switch status {
		case MonitorStatusOperational:
			return 3
		case MonitorStatusDegraded:
			return 2
		case MonitorStatusFailed, MonitorStatusError:
			return 1
		default:
			return 0
		}
	}
	for _, probe := range probes {
		if !probe.Enabled || probe.Status == "" {
			continue
		}
		candidate := &CheckResult{Status: probe.Status, LatencyMs: probe.LatencyMs}
		if best == nil || rank(candidate.Status) > rank(best.Status) ||
			(rank(candidate.Status) == rank(best.Status) && candidate.LatencyMs != nil && (best.LatencyMs == nil || *candidate.LatencyMs < *best.LatencyMs)) {
			best = candidate
		}
	}
	return best
}

// collectMonitorIndexes 把 monitors 列表按 ID 展开为聚合查询所需的三个索引结构。
func collectMonitorIndexes(monitors []*ChannelMonitor) ([]int64, map[int64]string, map[int64][]string, map[int64][]string) {
	ids := make([]int64, 0, len(monitors))
	primaryByID := make(map[int64]string, len(monitors))
	extrasByID := make(map[int64][]string, len(monitors))
	probeModelsByID := make(map[int64][]string, len(monitors))
	for _, m := range monitors {
		ids = append(ids, m.ID)
		primaryByID[m.ID] = m.PrimaryModel
		extrasByID[m.ID] = m.ExtraModels
		probeModelsByID[m.ID] = monitorProbeModels(m)
	}
	return ids, primaryByID, extrasByID, probeModelsByID
}

// monitorProbeModels 返回当前启用探针使用的模型，去重后保持配置顺序。
// 分组状态需要在这些模型之间取最佳结果，不能只看第一个探针。
func monitorProbeModels(m *ChannelMonitor) []string {
	if m == nil {
		return nil
	}
	if m.CheckMode == MonitorCheckModeQuota {
		if m.PrimaryModel == "" {
			return nil
		}
		return []string{m.PrimaryModel}
	}
	models := make([]string, 0, len(m.Probes))
	seen := make(map[string]struct{}, len(m.Probes))
	for _, probe := range m.Probes {
		if !probe.Enabled {
			continue
		}
		model := probeModel(probe, m.PrimaryModel)
		if model == "" {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		seen[model] = struct{}{}
		models = append(models, model)
	}
	if len(models) == 0 && m.PrimaryModel != "" {
		return []string{m.PrimaryModel}
	}
	return models
}

// batchLatest 批量取 latest per model，失败仅日志（与现有 BatchMonitorStatusSummary 一致，不阻断列表渲染）。
func (s *ChannelMonitorService) batchLatest(ctx context.Context, ids []int64) map[int64][]*ChannelMonitorLatest {
	latestMap, err := s.repo.ListLatestForMonitorIDs(ctx, ids)
	if err != nil {
		slog.Warn("channel_monitor: user view batch latest failed", "error", err)
		return map[int64][]*ChannelMonitorLatest{}
	}
	return latestMap
}

// batchTimeline 批量取每个 monitor 的分组聚合历史；旧版本未写聚合历史时回退到主模型历史。
func (s *ChannelMonitorService) batchTimeline(
	ctx context.Context,
	ids []int64,
	primaryByID map[int64]string,
) map[int64][]*ChannelMonitorHistoryEntry {
	timelineMap, err := s.repo.ListRecentHistoryForMonitors(ctx, ids, primaryByID, monitorTimelineMaxPoints)
	if err != nil {
		slog.Warn("channel_monitor: user view batch timeline failed", "error", err)
		return map[int64][]*ChannelMonitorHistoryEntry{}
	}
	for id, entries := range timelineMap {
		aggregated := make([]*ChannelMonitorHistoryEntry, 0, len(entries))
		legacy := make([]*ChannelMonitorHistoryEntry, 0, len(entries))
		for _, entry := range entries {
			if entry != nil && entry.Model == MonitorAggregateModel {
				aggregated = append(aggregated, entry)
			} else if entry != nil && entry.Model == primaryByID[id] {
				legacy = append(legacy, entry)
			}
		}
		if len(aggregated) > 0 {
			oldestAggregate := aggregated[len(aggregated)-1].CheckedAt
			olderLegacy := make([]*ChannelMonitorHistoryEntry, 0, len(legacy))
			for _, entry := range legacy {
				if entry.CheckedAt.Before(oldestAggregate) {
					olderLegacy = append(olderLegacy, entry)
				}
			}
			timelineMap[id] = mergeTimelineEntries(aggregated, olderLegacy, monitorTimelineMaxPoints)
			continue
		}
		timelineMap[id] = legacy
	}
	return timelineMap
}

// mergeTimelineEntries 合并新聚合历史和旧主模型历史，按时间倒序去重并限制返回数量。
func mergeTimelineEntries(aggregated, legacy []*ChannelMonitorHistoryEntry, limit int) []*ChannelMonitorHistoryEntry {
	merged := make([]*ChannelMonitorHistoryEntry, 0, len(aggregated)+len(legacy))
	seen := make(map[string]struct{}, len(aggregated)+len(legacy))
	for _, entry := range append(aggregated, legacy...) {
		if entry == nil {
			continue
		}
		key := entry.CheckedAt.UTC().Format(time.RFC3339Nano) + "|" + entry.Status
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		merged = append(merged, entry)
	}
	sort.SliceStable(merged, func(i, j int) bool { return merged[i].CheckedAt.After(merged[j].CheckedAt) })
	if limit > 0 && len(merged) > limit {
		merged = merged[:limit]
	}
	return merged
}

// pickLatest 从 latest 切片中挑出指定 model 对应项，未命中返回 nil。
func pickLatest(rows []*ChannelMonitorLatest, model string) *ChannelMonitorLatest {
	if model == "" {
		return nil
	}
	for _, r := range rows {
		if r.Model == model {
			return r
		}
	}
	return nil
}

// GetUserDetail 用户只读视图：单个监控详情（每个模型 7d/15d/30d 可用率与平均延迟）。
// 不暴露 api_key。
func (s *ChannelMonitorService) GetUserDetail(ctx context.Context, id int64) (*UserMonitorDetail, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !m.Enabled {
		return nil, ErrChannelMonitorNotFound
	}

	latest, err := s.repo.ListLatestPerModel(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list latest per model: %w", err)
	}
	availMap, err := s.collectAvailabilityWindows(ctx, id)
	if err != nil {
		return nil, err
	}

	models := mergeModelDetails(m, latest, availMap)
	return &UserMonitorDetail{
		ID:        m.ID,
		Name:      m.Name,
		Provider:  m.Provider,
		GroupName: m.GroupName,
		Models:    models,
	}, nil
}

// collectAvailabilityWindows 一次性查询 7/15/30 天三个窗口，按模型组织。
func (s *ChannelMonitorService) collectAvailabilityWindows(ctx context.Context, monitorID int64) (map[int]map[string]*ChannelMonitorAvailability, error) {
	out := make(map[int]map[string]*ChannelMonitorAvailability, 3)
	windows := []int{monitorAvailability7Days, monitorAvailability15Days, monitorAvailability30Days}
	for _, w := range windows {
		rows, err := s.repo.ComputeAvailability(ctx, monitorID, w)
		if err != nil {
			return nil, fmt.Errorf("compute availability %dd: %w", w, err)
		}
		out[w] = indexAvailabilityByModel(rows)
	}
	return out, nil
}

// ---------- 纯函数 helper（无 IO，可在 batch / 单 monitor / detail 路径复用）----------

// indexLatestByModel 把 latest 切片按 model 索引（小工具，避免在 hot path 重复写）。
func indexLatestByModel(rows []*ChannelMonitorLatest) map[string]*ChannelMonitorLatest {
	m := make(map[string]*ChannelMonitorLatest, len(rows))
	for _, r := range rows {
		m[r.Model] = r
	}
	return m
}

// indexAvailabilityByModel 把 availability 切片按 model 索引。
func indexAvailabilityByModel(rows []*ChannelMonitorAvailability) map[string]*ChannelMonitorAvailability {
	m := make(map[string]*ChannelMonitorAvailability, len(rows))
	for _, r := range rows {
		m[r.Model] = r
	}
	return m
}

// buildStatusSummary 由 latest + availability 字典构造 MonitorStatusSummary。
// 不做任何 IO，纯组装，便于在 batch 与单 monitor 路径复用。
func buildStatusSummary(
	latestByModel map[string]*ChannelMonitorLatest,
	availByModel map[string]*ChannelMonitorAvailability,
	primary string,
	extras []string,
	probeModels []string,
	latestRows ...[]*ChannelMonitorLatest,
) MonitorStatusSummary {
	summary := MonitorStatusSummary{ExtraModels: make([]ExtraModelStatus, 0, len(extras))}
	if primary != "" {
		latest := latestByModel[primary]
		if len(latestRows) > 0 {
			if best := bestProbeLatestRows(latestRows[0]); best != nil {
				latest = best
			}
		} else if best := bestProbeLatest(latestByModel, probeModels); best != nil {
			latest = best
		}
		if latest != nil {
			summary.PrimaryStatus = latest.Status
			summary.PrimaryLatencyMs = latest.LatencyMs
			// 配额快照挂在聚合行（quota / quota_probe 模式）或主模型行（旧记录）上。
			summary.LatestQuota = latest.Quota
		}
		if a, ok := availByModel[""]; ok {
			summary.Availability7d = a.AvailabilityPct
		} else if a, ok := availByModel[primary]; ok {
			summary.Availability7d = a.AvailabilityPct
		}
	}
	for _, model := range extras {
		entry := ExtraModelStatus{Model: model}
		if l, ok := latestByModel[model]; ok {
			entry.Status = l.Status
			entry.LatencyMs = l.LatencyMs
		}
		summary.ExtraModels = append(summary.ExtraModels, entry)
	}
	return summary
}

func bestProbeLatestRows(rows []*ChannelMonitorLatest) *ChannelMonitorLatest {
	for _, row := range rows {
		if row != nil && row.Model == MonitorAggregateModel {
			return row
		}
	}
	var best *ChannelMonitorLatest
	rank := func(status string) int {
		switch status {
		case MonitorStatusOperational:
			return 3
		case MonitorStatusDegraded:
			return 2
		case MonitorStatusFailed, MonitorStatusError:
			return 1
		default:
			return 0
		}
	}
	for _, candidate := range rows {
		if candidate == nil || candidate.Model == MonitorAggregateModel {
			continue
		}
		if best == nil || rank(candidate.Status) > rank(best.Status) ||
			(rank(candidate.Status) == rank(best.Status) && candidate.LatencyMs != nil && (best.LatencyMs == nil || *candidate.LatencyMs < *best.LatencyMs)) {
			best = candidate
		}
	}
	return best
}

// bestProbeLatest 按“任一探针正常即可用”的规则选择分组当前状态。
// 同一状态下优先返回延迟更低的探针，保证卡片延迟与状态来源一致。
func bestProbeLatest(latestByModel map[string]*ChannelMonitorLatest, models []string) *ChannelMonitorLatest {
	var best *ChannelMonitorLatest
	statusRank := func(status string) int {
		switch status {
		case "operational":
			return 3
		case "degraded":
			return 2
		case "failed":
			return 1
		case "error":
			return 0
		default:
			return -1
		}
	}
	for _, model := range models {
		candidate := latestByModel[model]
		if candidate == nil {
			continue
		}
		if best == nil || statusRank(candidate.Status) > statusRank(best.Status) ||
			(statusRank(candidate.Status) == statusRank(best.Status) && candidate.LatencyMs != nil && (best.LatencyMs == nil || *candidate.LatencyMs < *best.LatencyMs)) {
			best = candidate
		}
	}
	return best
}

// buildUserViewFromSummary 用预聚合好的 MonitorStatusSummary + 主模型 latest + timeline 装填 UserMonitorView（无 IO）。
// primaryLatest 可能为 nil（该监控尚无历史）；timelineEntries 可能为空。
func buildUserViewFromSummary(
	m *ChannelMonitor,
	summary MonitorStatusSummary,
	primaryLatest *ChannelMonitorLatest,
	timelineEntries []*ChannelMonitorHistoryEntry,
) *UserMonitorView {
	view := &UserMonitorView{
		ID:               m.ID,
		Name:             m.Name,
		Provider:         m.Provider,
		GroupName:        m.GroupName,
		PrimaryModel:     m.PrimaryModel,
		PrimaryStatus:    summary.PrimaryStatus,
		PrimaryLatencyMs: summary.PrimaryLatencyMs,
		Availability7d:   summary.Availability7d,
		ExtraModels:      summary.ExtraModels,
		Timeline:         buildTimelinePoints(timelineEntries),
	}
	if primaryLatest != nil {
		view.PrimaryPingLatencyMs = primaryLatest.PingLatencyMs
		view.LatestQuota = primaryLatest.Quota
	}
	return view
}

// buildTimelinePoints 把 history entry 裁剪为 timeline 点（去除 message/ID/Model，减小响应体）。
func buildTimelinePoints(entries []*ChannelMonitorHistoryEntry) []UserMonitorTimelinePoint {
	out := make([]UserMonitorTimelinePoint, 0, len(entries))
	for _, e := range entries {
		out = append(out, UserMonitorTimelinePoint{
			Status:        e.Status,
			LatencyMs:     e.LatencyMs,
			PingLatencyMs: e.PingLatencyMs,
			CheckedAt:     e.CheckedAt,
		})
	}
	return out
}

// mergeModelDetails 合并 latest + availability 三个窗口为 ModelDetail 列表。
// 复用 indexLatestByModel，避免在多处重复写 build map 逻辑。
func mergeModelDetails(
	m *ChannelMonitor,
	latest []*ChannelMonitorLatest,
	availMap map[int]map[string]*ChannelMonitorAvailability,
) []ModelDetail {
	latestByModel := indexLatestByModel(latest)
	if aggregate, ok := latestByModel[MonitorAggregateModel]; ok {
		d := ModelDetail{Model: m.PrimaryModel, LatestStatus: aggregate.Status, LatestLatencyMs: aggregate.LatencyMs}
		for _, days := range []int{monitorAvailability7Days, monitorAvailability15Days, monitorAvailability30Days} {
			if a, ok := availMap[days][""]; ok {
				switch days {
				case monitorAvailability7Days:
					d.Availability7d, d.AvgLatency7dMs = a.AvailabilityPct, a.AvgLatencyMs
				case monitorAvailability15Days:
					d.Availability15d = a.AvailabilityPct
				case monitorAvailability30Days:
					d.Availability30d = a.AvailabilityPct
				}
			}
		}
		return []ModelDetail{d}
	}
	all := append([]string{m.PrimaryModel}, m.ExtraModels...)
	out := make([]ModelDetail, 0, len(all))
	for _, model := range all {
		d := ModelDetail{Model: model}
		if l, ok := latestByModel[model]; ok {
			d.LatestStatus = l.Status
			d.LatestLatencyMs = l.LatencyMs
		}
		if a, ok := availMap[monitorAvailability7Days][model]; ok {
			d.Availability7d = a.AvailabilityPct
			d.AvgLatency7dMs = a.AvgLatencyMs
		}
		if a, ok := availMap[monitorAvailability15Days][model]; ok {
			d.Availability15d = a.AvailabilityPct
		}
		if a, ok := availMap[monitorAvailability30Days][model]; ok {
			d.Availability30d = a.AvailabilityPct
		}
		out = append(out, d)
	}
	return out
}
