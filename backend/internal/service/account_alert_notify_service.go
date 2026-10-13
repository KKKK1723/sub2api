package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/google/uuid"
)

const (
	// accountAlertNotifyLeaderLockKey 保证蓝绿切换等多实例并存时只有一个实例做检查和发信。
	accountAlertNotifyLeaderLockKey = "account:alert:notify:leader"
	// accountAlertNotifyLeaderLockTTL 只用于崩溃兜底，需大于单轮最坏耗时（含多次 SMTP 超时）。
	accountAlertNotifyLeaderLockTTL = 3 * time.Minute
	accountAlertNotifyCycleTimeout  = 2 * time.Minute
	accountAlertNotifyPageSize      = 100
	accountAlertNotifyMaxPages      = 1000
	accountAlertNotifyStateVersion  = 1
)

// accountAlertGroupState 记录一个分组「可用账号为 0」的持续情况。
type accountAlertGroupState struct {
	ZeroSince time.Time `json:"zero_since"`
	Notified  bool      `json:"notified"`
	// Silent 表示首次检查（基线）时分组已不可用：不发不可用通知，恢复时也不发。
	Silent bool `json:"silent,omitempty"`
}

// accountAlertNotifyState 后台任务的运行状态，存在设置表中供多实例共享，只在变化时写入。
type accountAlertNotifyState struct {
	Version             int                               `json:"version"`
	AccountsInitialized bool                              `json:"accounts_initialized"`
	ErrorAccountIDs     []int64                           `json:"error_account_ids"`
	GroupsInitialized   bool                              `json:"groups_initialized"`
	Groups              map[int64]*accountAlertGroupState `json:"groups"`
	LowBalanceHosts     []string                          `json:"low_balance_hosts"`
}

func (st *accountAlertNotifyState) isEmpty() bool {
	return !st.AccountsInitialized && len(st.ErrorAccountIDs) == 0 &&
		!st.GroupsInitialized && len(st.Groups) == 0 && len(st.LowBalanceHosts) == 0
}

// AccountAlertNotifyService 定时检查账号、分组可用性和上游余额，并按配置发送管理员通知邮件。
type AccountAlertNotifyService struct {
	accountRepo  AccountRepository
	settingRepo  SettingRepository
	emailService *EmailService
	interval     time.Duration
	stopCh       chan struct{}
	stopOnce     sync.Once
	wg           sync.WaitGroup

	lockCache  LeaderLockCache
	db         *sql.DB
	instanceID string
}

func NewAccountAlertNotifyService(accountRepo AccountRepository, settingRepo SettingRepository, emailService *EmailService, interval time.Duration) *AccountAlertNotifyService {
	return &AccountAlertNotifyService{
		accountRepo:  accountRepo,
		settingRepo:  settingRepo,
		emailService: emailService,
		interval:     interval,
		stopCh:       make(chan struct{}),
		instanceID:   uuid.NewString(),
	}
}

// SetLeaderLock 注入分布式锁；两者都为 nil 时不加锁运行（单实例/测试场景）。
func (s *AccountAlertNotifyService) SetLeaderLock(lockCache LeaderLockCache, db *sql.DB) {
	if s == nil {
		return
	}
	s.lockCache = lockCache
	s.db = db
}

func (s *AccountAlertNotifyService) Start() {
	if s == nil || s.accountRepo == nil || s.settingRepo == nil || s.emailService == nil || s.interval <= 0 {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		s.runOnce()
		for {
			select {
			case <-ticker.C:
				s.runOnce()
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *AccountAlertNotifyService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
	s.wg.Wait()
}

func (s *AccountAlertNotifyService) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), accountAlertNotifyCycleTimeout)
	defer cancel()

	cfg, err := loadAccountAlertNotifyConfig(ctx, s.settingRepo)
	if err != nil {
		slog.Warn("account_alert_notify.load_config_failed", "error", err)
		return
	}

	release, ok := tryAcquireSingletonLeaderLock(ctx, s.lockCache, s.db, accountAlertNotifyLeaderLockKey, s.instanceID, accountAlertNotifyLeaderLockTTL)
	if !ok {
		return
	}
	defer release()

	state, err := s.loadState(ctx)
	if err != nil {
		slog.Warn("account_alert_notify.load_state_failed", "error", err)
		return
	}
	before, _ := json.Marshal(state)

	recipients := cfg.Recipients()
	active := len(recipients) > 0 && (cfg.AccountErrorEnabled || cfg.GroupUnavailableEnabled || cfg.UpstreamBalanceLowEnabled)
	if !active {
		// 没有收件人或全部关闭：清空状态，重新开启后先记录现状，不补发历史变更。
		if !state.isEmpty() {
			s.saveState(ctx, &accountAlertNotifyState{Version: accountAlertNotifyStateVersion})
		}
		return
	}

	accounts, err := s.listAllAccounts(ctx)
	if err != nil {
		slog.Warn("account_alert_notify.list_accounts_failed", "error", err)
		return
	}

	now := time.Now()
	siteName := s.siteName(ctx)
	adminURL := s.adminURL(ctx)

	s.checkAccountErrors(ctx, cfg, recipients, accounts, state, now, siteName, adminURL)
	s.checkGroups(ctx, cfg, recipients, accounts, state, now, siteName, adminURL)
	s.checkUpstreamBalances(ctx, cfg, recipients, accounts, state, siteName)

	state.Version = accountAlertNotifyStateVersion
	after, _ := json.Marshal(state)
	if string(before) != string(after) {
		s.saveState(ctx, state)
	}
}

// checkAccountErrors 账号从非错误变为错误时通知；首次检查只记录现状。
func (s *AccountAlertNotifyService) checkAccountErrors(ctx context.Context, cfg *AccountAlertNotifyConfig, recipients []string, accounts []Account, state *accountAlertNotifyState, now time.Time, siteName, adminURL string) {
	if !cfg.AccountErrorEnabled {
		state.AccountsInitialized = false
		state.ErrorAccountIDs = nil
		return
	}

	current := make([]*Account, 0)
	for i := range accounts {
		if accounts[i].Status == StatusError {
			current = append(current, &accounts[i])
		}
	}

	if !state.AccountsInitialized {
		state.ErrorAccountIDs = accountAlertAccountIDs(current)
		state.AccountsInitialized = true
		return
	}

	previous := make(map[int64]struct{}, len(state.ErrorAccountIDs))
	for _, id := range state.ErrorAccountIDs {
		previous[id] = struct{}{}
	}
	var newly []*Account
	for _, acc := range current {
		if _, ok := previous[acc.ID]; !ok {
			newly = append(newly, acc)
		}
	}

	sent := true
	if len(newly) > 0 {
		rows := make([]accountAlertErrorRow, 0, len(newly))
		for _, acc := range newly {
			rows = append(rows, accountAlertErrorRow{
				ID:       acc.ID,
				Name:     acc.Name,
				Platform: acc.Platform,
				Type:     acc.Type,
				Groups:   accountAlertGroupNames(acc),
				Reason:   truncateAccountAlertText(acc.ErrorMessage, accountAlertReasonMaxRunes),
			})
		}
		subject, body := buildAccountErrorAlertEmail(siteName, adminURL, rows, now)
		sent = s.send(ctx, recipients, subject, body)
	}

	// 发送失败时不记录本轮新出错的账号，下一轮重试；已恢复的账号自然从集合中移除。
	next := make([]*Account, 0, len(current))
	for _, acc := range current {
		if _, ok := previous[acc.ID]; ok || sent {
			next = append(next, acc)
		}
	}
	state.ErrorAccountIDs = accountAlertAccountIDs(next)
}

type accountAlertGroupInfo struct {
	group   *Group
	members []*Account
}

// checkGroups 分组可用账号数持续为 0 超过设定时长时通知，恢复后按配置发送恢复通知。
func (s *AccountAlertNotifyService) checkGroups(ctx context.Context, cfg *AccountAlertNotifyConfig, recipients []string, accounts []Account, state *accountAlertNotifyState, now time.Time, siteName, adminURL string) {
	if !cfg.GroupUnavailableEnabled {
		state.GroupsInitialized = false
		state.Groups = nil
		return
	}

	groups := make(map[int64]*accountAlertGroupInfo)
	for i := range accounts {
		acc := &accounts[i]
		for _, grp := range acc.Groups {
			if grp == nil || grp.ID <= 0 {
				continue
			}
			info, ok := groups[grp.ID]
			if !ok {
				info = &accountAlertGroupInfo{group: grp}
				groups[grp.ID] = info
			}
			info.members = append(info.members, acc)
		}
	}

	baseline := !state.GroupsInitialized
	if state.Groups == nil {
		state.Groups = make(map[int64]*accountAlertGroupState)
	}
	delay := time.Duration(cfg.GroupUnavailableDelayMinutes) * time.Minute

	var downItems []accountAlertGroupDownItem
	var downIDs []int64
	var upItems []accountAlertGroupUpItem
	var upIDs []int64
	monitored := make(map[int64]struct{}, len(groups))

	groupIDs := make([]int64, 0, len(groups))
	for id := range groups {
		groupIDs = append(groupIDs, id)
	}
	sort.Slice(groupIDs, func(i, j int) bool { return groupIDs[i] < groupIDs[j] })

	for _, gid := range groupIDs {
		info := groups[gid]
		// 只监控启用中且至少绑定了 1 个账号的分组
		if !info.group.IsActive() || len(info.members) == 0 {
			continue
		}
		monitored[gid] = struct{}{}

		available := 0
		for _, acc := range info.members {
			if acc.IsSchedulable() {
				available++
			}
		}

		gs := state.Groups[gid]
		if available == 0 {
			if gs == nil {
				gs = &accountAlertGroupState{ZeroSince: now}
				if baseline {
					gs.Notified = true
					gs.Silent = true
				}
				state.Groups[gid] = gs
			}
			if !gs.Notified && now.Sub(gs.ZeroSince) >= delay {
				downItems = append(downItems, buildAccountAlertGroupDownItem(info, gs.ZeroSince, now))
				downIDs = append(downIDs, gid)
			}
			continue
		}

		if gs == nil {
			continue
		}
		if gs.Notified && !gs.Silent && cfg.GroupRecoveredEnabled {
			upItems = append(upItems, buildAccountAlertGroupUpItem(info, gs.ZeroSince, now, available))
			upIDs = append(upIDs, gid)
			continue
		}
		delete(state.Groups, gid)
	}

	// 分组被停用、删除或不再绑定账号时，丢弃其状态，不发通知。
	for gid := range state.Groups {
		if _, ok := monitored[gid]; !ok {
			delete(state.Groups, gid)
		}
	}

	if len(downItems) > 0 {
		subject, body := buildGroupUnavailableAlertEmail(siteName, adminURL, downItems, now)
		if s.send(ctx, recipients, subject, body) {
			for _, gid := range downIDs {
				state.Groups[gid].Notified = true
			}
		}
	}
	if len(upItems) > 0 {
		subject, body := buildGroupRecoveredAlertEmail(siteName, upItems)
		// 恢复通知发送失败时保留状态，下一轮重试
		if s.send(ctx, recipients, subject, body) {
			for _, gid := range upIDs {
				delete(state.Groups, gid)
			}
		}
	}
	state.GroupsInitialized = true
}

func buildAccountAlertGroupDownItem(info *accountAlertGroupInfo, zeroSince, now time.Time) accountAlertGroupDownItem {
	item := accountAlertGroupDownItem{
		ID:        info.group.ID,
		Name:      info.group.Name,
		Platform:  info.group.Platform,
		Total:     len(info.members),
		ZeroSince: zeroSince,
		Members:   make([]accountAlertGroupMember, 0, len(info.members)),
	}
	for _, acc := range info.members {
		label, tagClass, detail, recovery := describeAccountAlertAvailability(acc, now)
		item.Members = append(item.Members, accountAlertGroupMember{
			ID:       acc.ID,
			Name:     acc.Name,
			Label:    label,
			TagClass: tagClass,
			Detail:   detail,
			Recovery: recovery,
		})
	}
	return item
}

func buildAccountAlertGroupUpItem(info *accountAlertGroupInfo, zeroSince, now time.Time, available int) accountAlertGroupUpItem {
	item := accountAlertGroupUpItem{
		ID:          info.group.ID,
		Name:        info.group.Name,
		Platform:    info.group.Platform,
		Available:   available,
		Total:       len(info.members),
		ZeroSince:   zeroSince,
		RecoveredAt: now,
	}
	for _, acc := range info.members {
		if acc.IsSchedulable() {
			item.Members = append(item.Members, accountAlertAccountLabel(acc.ID, acc.Name))
		}
	}
	return item
}

// describeAccountAlertAvailability 说明账号当前是否可调度及原因，判断顺序与 Account.IsSchedulable 保持一致。
func describeAccountAlertAvailability(acc *Account, now time.Time) (label, tagClass, detail, recovery string) {
	switch {
	case acc.Status == StatusError:
		return "错误", "tag-error", truncateAccountAlertText(acc.ErrorMessage, accountAlertReasonMaxRunes), "需手动恢复"
	case !acc.IsActive():
		return "已停用", "tag-error", "", "需手动启用"
	case !acc.Schedulable:
		return "已关闭调度", "tag-error", "", "需手动开启调度"
	case acc.AutoPauseOnExpired && acc.ExpiresAt != nil && !now.Before(*acc.ExpiresAt):
		return "已过期", "tag-error", "到期时间 " + acc.ExpiresAt.Local().Format(accountAlertTimeLayout), "需续期"
	case acc.OverloadUntil != nil && now.Before(*acc.OverloadUntil):
		return "过载", "tag-limit", "", "预计 " + formatAccountAlertUntil(*acc.OverloadUntil, now) + " 自动恢复"
	case acc.RateLimitResetAt != nil && now.Before(*acc.RateLimitResetAt):
		return "限流", "tag-limit", "", "预计 " + formatAccountAlertUntil(*acc.RateLimitResetAt, now) + " 自动恢复"
	case acc.TempUnschedulableUntil != nil && now.Before(*acc.TempUnschedulableUntil):
		return "临时不可调度", "tag-temp", accountAlertTempUnschedReason(acc.TempUnschedulableReason), "预计 " + formatAccountAlertUntil(*acc.TempUnschedulableUntil, now) + " 自动恢复"
	case acc.IsAPIKeyOrBedrock() && acc.IsQuotaExceeded():
		return "额度已用完", "tag-limit", "", "额度重置后恢复"
	default:
		return "可用", "tag-ok", "", ""
	}
}

// accountAlertTempUnschedReason 临时不可调度原因可能是 JSON（如流超时规则），优先取其中的错误信息。
func accountAlertTempUnschedReason(reason string) string {
	reason = strings.TrimSpace(reason)
	if strings.HasPrefix(reason, "{") {
		var payload map[string]any
		if err := json.Unmarshal([]byte(reason), &payload); err == nil {
			for _, key := range []string{"error_message", "reason", "message"} {
				if v, ok := payload[key].(string); ok && strings.TrimSpace(v) != "" {
					reason = v
					break
				}
			}
		}
	}
	return truncateAccountAlertText(reason, accountAlertReasonMaxRunes)
}

func formatAccountAlertUntil(t, now time.Time) string {
	local := t.Local()
	nowLocal := now.Local()
	if local.Year() == nowLocal.Year() && local.YearDay() == nowLocal.YearDay() {
		return local.Format("15:04")
	}
	return local.Format("01-02 15:04")
}

type accountAlertBalanceHostAgg struct {
	accounts []accountAlertBalanceAccount
	min      float64
	unit     string
	hasData  bool
}

// checkUpstreamBalances 对开启余额探测的上游按域名聚合，最低余额刚跌破阈值时通知一次。
func (s *AccountAlertNotifyService) checkUpstreamBalances(ctx context.Context, cfg *AccountAlertNotifyConfig, recipients []string, accounts []Account, state *accountAlertNotifyState, siteName string) {
	if !cfg.UpstreamBalanceLowEnabled {
		state.LowBalanceHosts = nil
		return
	}

	now := time.Now()
	threshold := cfg.UpstreamBalanceLowThreshold
	hosts := make(map[string]*accountAlertBalanceHostAgg)
	for i := range accounts {
		acc := &accounts[i]
		if !isUpstreamBalanceProbeAccount(acc) || !upstreamBalanceProbeEnabled(acc) {
			continue
		}
		host := accountAlertUpstreamHost(acc)
		if host == "" {
			continue
		}
		agg, ok := hosts[host]
		if !ok {
			agg = &accountAlertBalanceHostAgg{}
			hosts[host] = agg
		}
		snapshot := decodeUpstreamBalanceProbeSnapshot(acc.Extra)
		if !accountAlertBalanceSnapshotUsable(snapshot, now) {
			continue
		}
		balance := snapshot.Data.Balance
		agg.accounts = append(agg.accounts, accountAlertBalanceAccount{
			ID:         acc.ID,
			Name:       acc.Name,
			Balance:    balance,
			Unit:       snapshot.Data.Unit,
			ReceivedAt: snapshot.ReceivedAt,
			Low:        balance < threshold,
		})
		if !agg.hasData || balance < agg.min {
			agg.min = balance
			agg.unit = snapshot.Data.Unit
		}
		agg.hasData = true
	}

	previousLow := make(map[string]struct{}, len(state.LowBalanceHosts))
	for _, host := range state.LowBalanceHosts {
		previousLow[host] = struct{}{}
	}

	lowNow := make(map[string]struct{})
	var notifyItems []accountAlertBalanceHostItem
	var notifyHosts []string
	hostNames := make([]string, 0, len(hosts))
	for host := range hosts {
		hostNames = append(hostNames, host)
	}
	sort.Strings(hostNames)
	for _, host := range hostNames {
		agg := hosts[host]
		if !agg.hasData {
			// 暂时查不到余额时保持原状态，避免探测失败恢复后重复通知
			if _, ok := previousLow[host]; ok {
				lowNow[host] = struct{}{}
			}
			continue
		}
		if agg.min >= threshold {
			continue
		}
		lowNow[host] = struct{}{}
		if _, ok := previousLow[host]; ok {
			continue
		}
		sort.Slice(agg.accounts, func(i, j int) bool { return agg.accounts[i].Balance < agg.accounts[j].Balance })
		notifyItems = append(notifyItems, accountAlertBalanceHostItem{
			Host:       host,
			MinBalance: agg.min,
			Unit:       agg.unit,
			Accounts:   agg.accounts,
		})
		notifyHosts = append(notifyHosts, host)
	}

	if len(notifyItems) > 0 {
		subject, body := buildUpstreamBalanceLowAlertEmail(siteName, threshold, notifyItems)
		if !s.send(ctx, recipients, subject, body) {
			// 发送失败不记为已通知，下一轮重试
			for _, host := range notifyHosts {
				delete(lowNow, host)
			}
		}
	}

	next := make([]string, 0, len(lowNow))
	for host := range lowNow {
		next = append(next, host)
	}
	sort.Strings(next)
	state.LowBalanceHosts = next
}

// accountAlertBalanceSnapshotUsable 只使用成功的探测结果；探测失败但仍在有效期内的旧数据也可用。
func accountAlertBalanceSnapshotUsable(snapshot *UpstreamBalanceProbeSnapshot, now time.Time) bool {
	if snapshot == nil || snapshot.Data == nil {
		return false
	}
	if snapshot.Data.IsValid != nil && !*snapshot.Data.IsValid {
		return false
	}
	if math.IsNaN(snapshot.Data.Balance) || math.IsInf(snapshot.Data.Balance, 0) {
		return false
	}
	if snapshot.Status == UpstreamBalanceProbeStatusOK {
		return true
	}
	return snapshot.FreshUntil != nil && now.Before(*snapshot.FreshUntil)
}

func accountAlertUpstreamHost(acc *Account) string {
	baseURL := UpstreamBalanceBaseURL(acc)
	if baseURL == "" {
		return ""
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	return strings.ToLower(parsed.Host)
}

func accountAlertAccountIDs(accounts []*Account) []int64 {
	ids := make([]int64, 0, len(accounts))
	for _, acc := range accounts {
		ids = append(ids, acc.ID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func accountAlertGroupNames(acc *Account) []string {
	names := make([]string, 0, len(acc.Groups))
	for _, grp := range acc.Groups {
		if grp != nil && strings.TrimSpace(grp.Name) != "" {
			names = append(names, grp.Name)
		}
	}
	return names
}

// send 逐个收件人发送，至少一个成功即视为已通知。
func (s *AccountAlertNotifyService) send(ctx context.Context, recipients []string, subject, body string) bool {
	delivered := false
	for _, to := range recipients {
		// SMTP 读写有自己的超时，不受 ctx 控制；本轮超时后不再继续发，避免占锁过久
		if ctx.Err() != nil {
			break
		}
		if err := s.emailService.SendEmail(ctx, to, subject, body); err != nil {
			slog.Warn("account_alert_notify.send_failed", "to", to, "subject", subject, "error", err)
			continue
		}
		delivered = true
	}
	if delivered {
		slog.Info("account_alert_notify.sent", "subject", subject, "recipients", len(recipients))
	}
	return delivered
}

func (s *AccountAlertNotifyService) listAllAccounts(ctx context.Context) ([]Account, error) {
	out := make([]Account, 0, accountAlertNotifyPageSize)
	for page := 1; page <= accountAlertNotifyMaxPages; page++ {
		accounts, pageInfo, err := s.accountRepo.ListWithFilters(ctx, pagination.PaginationParams{
			Page:     page,
			PageSize: accountAlertNotifyPageSize,
		}, "", "", "", "", 0, "")
		if err != nil {
			return nil, err
		}
		out = append(out, accounts...)
		if len(accounts) < accountAlertNotifyPageSize {
			break
		}
		if pageInfo != nil && int64(len(out)) >= pageInfo.Total {
			break
		}
	}
	return out, nil
}

func (s *AccountAlertNotifyService) loadState(ctx context.Context) (*accountAlertNotifyState, error) {
	state := &accountAlertNotifyState{Version: accountAlertNotifyStateVersion}
	value, err := s.settingRepo.GetValue(ctx, SettingKeyAccountAlertNotifyState)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return state, nil
		}
		return nil, err
	}
	if strings.TrimSpace(value) == "" {
		return state, nil
	}
	if err := json.Unmarshal([]byte(value), state); err != nil {
		// 状态损坏时重新建立基线，宁可漏报一次也不批量误报
		slog.Warn("account_alert_notify.decode_state_failed", "error", err)
		return &accountAlertNotifyState{Version: accountAlertNotifyStateVersion}, nil
	}
	return state, nil
}

func (s *AccountAlertNotifyService) saveState(ctx context.Context, state *accountAlertNotifyState) {
	data, err := json.Marshal(state)
	if err != nil {
		slog.Warn("account_alert_notify.encode_state_failed", "error", err)
		return
	}
	if err := s.settingRepo.Set(ctx, SettingKeyAccountAlertNotifyState, string(data)); err != nil {
		slog.Warn("account_alert_notify.save_state_failed", "error", err)
	}
}

func (s *AccountAlertNotifyService) siteName(ctx context.Context) string {
	name, err := s.settingRepo.GetValue(ctx, SettingKeySiteName)
	if err != nil || strings.TrimSpace(name) == "" {
		return defaultSiteName
	}
	return strings.TrimSpace(name)
}

// adminURL 邮件中「前往账号管理」按钮的站点地址，优先使用前端地址。
func (s *AccountAlertNotifyService) adminURL(ctx context.Context) string {
	for _, key := range []string{SettingKeyFrontendURL, SettingKeyAPIBaseURL} {
		value, err := s.settingRepo.GetValue(ctx, key)
		if err == nil && strings.TrimSpace(value) != "" {
			return strings.TrimRight(strings.TrimSpace(value), "/")
		}
	}
	return ""
}
