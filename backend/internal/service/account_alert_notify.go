package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"math"
	"net/mail"
	"strconv"
	"strings"
	"time"
)

// 账号与上游监控通知：账号变为错误、分组无可用账号、上游余额不足时给管理员发邮件。
// 设置项键名放在本功能文件内，避免改动与上游共享的 domain_constants.go，降低同步冲突。
const (
	SettingKeyAccountAlertNotifyConfig = "account_alert_notify_config" // 通知配置（JSON）
	SettingKeyAccountAlertNotifyState  = "account_alert_notify_state"  // 运行状态（JSON，仅后台任务读写）

	accountAlertGroupDelayDefaultMinutes = 3
	accountAlertGroupDelayMaxMinutes     = 60
	accountAlertBalanceThresholdDefault  = 10.0
	accountAlertBalanceThresholdMax      = 1e9
	accountAlertMaxRecipients            = 10
	accountAlertReasonMaxRunes           = 300
	accountAlertTimeLayout               = "2006-01-02 15:04:05"
)

// AccountAlertNotifyConfig 账号与上游监控通知配置
type AccountAlertNotifyConfig struct {
	Emails                       []NotifyEmailEntry `json:"emails"`
	AccountErrorEnabled          bool               `json:"account_error_enabled"`
	GroupUnavailableEnabled      bool               `json:"group_unavailable_enabled"`
	GroupUnavailableDelayMinutes int                `json:"group_unavailable_delay_minutes"`
	GroupRecoveredEnabled        bool               `json:"group_recovered_enabled"`
	UpstreamBalanceLowEnabled    bool               `json:"upstream_balance_low_enabled"`
	UpstreamBalanceLowThreshold  float64            `json:"upstream_balance_low_threshold"`
}

// DefaultAccountAlertNotifyConfig 返回默认配置：三类通知默认开启，收件邮箱为空时不发送。
func DefaultAccountAlertNotifyConfig() *AccountAlertNotifyConfig {
	return &AccountAlertNotifyConfig{
		Emails:                       []NotifyEmailEntry{},
		AccountErrorEnabled:          true,
		GroupUnavailableEnabled:      true,
		GroupUnavailableDelayMinutes: accountAlertGroupDelayDefaultMinutes,
		GroupRecoveredEnabled:        true,
		UpstreamBalanceLowEnabled:    true,
		UpstreamBalanceLowThreshold:  accountAlertBalanceThresholdDefault,
	}
}

// Recipients 返回启用中的收件邮箱（去重，保持原顺序）。
func (c *AccountAlertNotifyConfig) Recipients() []string {
	if c == nil {
		return nil
	}
	recipients := make([]string, 0, len(c.Emails))
	seen := make(map[string]struct{}, len(c.Emails))
	for _, entry := range c.Emails {
		email := strings.TrimSpace(entry.Email)
		if entry.Disabled || email == "" {
			continue
		}
		lower := strings.ToLower(email)
		if _, ok := seen[lower]; ok {
			continue
		}
		seen[lower] = struct{}{}
		recipients = append(recipients, email)
	}
	return recipients
}

// normalizeAccountAlertNotifyConfig 修正读取到的配置，保证后台任务拿到的值都在合法范围内。
func normalizeAccountAlertNotifyConfig(cfg *AccountAlertNotifyConfig) {
	if cfg.GroupUnavailableDelayMinutes < 0 {
		cfg.GroupUnavailableDelayMinutes = 0
	}
	if cfg.GroupUnavailableDelayMinutes > accountAlertGroupDelayMaxMinutes {
		cfg.GroupUnavailableDelayMinutes = accountAlertGroupDelayMaxMinutes
	}
	if math.IsNaN(cfg.UpstreamBalanceLowThreshold) || math.IsInf(cfg.UpstreamBalanceLowThreshold, 0) || cfg.UpstreamBalanceLowThreshold < 0 {
		cfg.UpstreamBalanceLowThreshold = accountAlertBalanceThresholdDefault
	}
	emails := make([]NotifyEmailEntry, 0, len(cfg.Emails))
	seen := make(map[string]struct{}, len(cfg.Emails))
	for _, entry := range cfg.Emails {
		entry.Email = strings.TrimSpace(entry.Email)
		if entry.Email == "" {
			continue
		}
		lower := strings.ToLower(entry.Email)
		if _, ok := seen[lower]; ok {
			continue
		}
		seen[lower] = struct{}{}
		// 管理员自己填写的收件邮箱不走验证码流程
		entry.Verified = true
		emails = append(emails, entry)
		if len(emails) >= accountAlertMaxRecipients {
			break
		}
	}
	cfg.Emails = emails
}

func validateAccountAlertNotifyConfig(cfg *AccountAlertNotifyConfig) error {
	if cfg == nil {
		return errors.New("settings cannot be nil")
	}
	if cfg.GroupUnavailableDelayMinutes < 0 || cfg.GroupUnavailableDelayMinutes > accountAlertGroupDelayMaxMinutes {
		return fmt.Errorf("group_unavailable_delay_minutes must be between 0-%d", accountAlertGroupDelayMaxMinutes)
	}
	if math.IsNaN(cfg.UpstreamBalanceLowThreshold) || math.IsInf(cfg.UpstreamBalanceLowThreshold, 0) ||
		cfg.UpstreamBalanceLowThreshold < 0 || cfg.UpstreamBalanceLowThreshold > accountAlertBalanceThresholdMax {
		return errors.New("upstream_balance_low_threshold is invalid")
	}
	count := 0
	for _, entry := range cfg.Emails {
		email := strings.TrimSpace(entry.Email)
		if email == "" {
			continue
		}
		if _, err := mail.ParseAddress(email); err != nil || strings.ContainsAny(email, "<> ") {
			return fmt.Errorf("invalid email: %s", email)
		}
		count++
	}
	if count > accountAlertMaxRecipients {
		return fmt.Errorf("at most %d emails are allowed", accountAlertMaxRecipients)
	}
	return nil
}

// loadAccountAlertNotifyConfig 读取配置；未保存过时返回默认值，存量 JSON 缺字段时用默认值补齐。
func loadAccountAlertNotifyConfig(ctx context.Context, repo SettingRepository) (*AccountAlertNotifyConfig, error) {
	cfg := DefaultAccountAlertNotifyConfig()
	if repo == nil {
		return cfg, nil
	}
	value, err := repo.GetValue(ctx, SettingKeyAccountAlertNotifyConfig)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return cfg, nil
		}
		return nil, fmt.Errorf("get account alert notify settings: %w", err)
	}
	if strings.TrimSpace(value) == "" {
		return cfg, nil
	}
	if err := json.Unmarshal([]byte(value), cfg); err != nil {
		return DefaultAccountAlertNotifyConfig(), nil
	}
	if cfg.Emails == nil {
		cfg.Emails = []NotifyEmailEntry{}
	}
	normalizeAccountAlertNotifyConfig(cfg)
	return cfg, nil
}

// GetAccountAlertNotifyConfig 获取账号与上游监控通知配置
func (s *SettingService) GetAccountAlertNotifyConfig(ctx context.Context) (*AccountAlertNotifyConfig, error) {
	return loadAccountAlertNotifyConfig(ctx, s.settingRepo)
}

// SetAccountAlertNotifyConfig 保存账号与上游监控通知配置
func (s *SettingService) SetAccountAlertNotifyConfig(ctx context.Context, cfg *AccountAlertNotifyConfig) error {
	if err := validateAccountAlertNotifyConfig(cfg); err != nil {
		return err
	}
	normalizeAccountAlertNotifyConfig(cfg)
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal account alert notify settings: %w", err)
	}
	return s.settingRepo.Set(ctx, SettingKeyAccountAlertNotifyConfig, string(data))
}

// ---------------------------------------------------------------------------
// 邮件内容
// ---------------------------------------------------------------------------

type accountAlertErrorRow struct {
	ID       int64
	Name     string
	Platform string
	Type     string
	Groups   []string
	Reason   string
}

type accountAlertGroupMember struct {
	ID       int64
	Name     string
	Label    string
	TagClass string
	Detail   string
	Recovery string
}

type accountAlertGroupDownItem struct {
	ID        int64
	Name      string
	Platform  string
	Total     int
	ZeroSince time.Time
	Members   []accountAlertGroupMember
}

type accountAlertGroupUpItem struct {
	ID          int64
	Name        string
	Platform    string
	Available   int
	Total       int
	ZeroSince   time.Time
	RecoveredAt time.Time
	Members     []string
}

type accountAlertBalanceAccount struct {
	ID         int64
	Name       string
	Balance    float64
	Unit       string
	ReceivedAt *time.Time
	Low        bool
}

type accountAlertBalanceHostItem struct {
	Host       string
	MinBalance float64
	Unit       string
	Accounts   []accountAlertBalanceAccount
}

// 样式沿用现有通知邮件模板；通过字符串拼接而非 Sprintf 嵌入，避免 CSS 中的 % 被当作格式符。
const accountAlertEmailStyle = `
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #f5f5f5; margin: 0; padding: 20px; color: #333; }
        .container { max-width: 640px; margin: 0 auto; background-color: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        .header { color: white; padding: 26px 30px; text-align: center; }
        .header h1 { margin: 0; font-size: 22px; }
        .header p { margin: 6px 0 0; font-size: 13px; opacity: 0.9; }
        .red { background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%); }
        .amber { background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%); }
        .green { background: linear-gradient(135deg, #10b981 0%, #059669 100%); }
        .blue { background: linear-gradient(135deg, #14b8a6 0%, #0d9488 100%); }
        .content { padding: 28px 30px; }
        .lead { font-size: 16px; color: #333; text-align: center; margin: 0 0 18px; }
        .metric { display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid #eee; font-size: 14px; }
        .metric-label { color: #666; }
        .metric-value { font-weight: bold; color: #333; text-align: right; }
        h4 { margin: 22px 0 8px; font-size: 14px; color: #333; }
        table { width: 100%; border-collapse: collapse; font-size: 13px; }
        th { text-align: left; padding: 8px 10px; background: #f8f9fa; color: #666; font-weight: 600; border-bottom: 1px solid #eee; white-space: nowrap; }
        td { padding: 9px 10px; border-bottom: 1px solid #f0f0f0; vertical-align: top; }
        .muted { color: #999; font-size: 12px; }
        .reason { color: #b91c1c; word-break: break-all; }
        .tag { display: inline-block; padding: 1px 8px; border-radius: 999px; font-size: 12px; font-weight: 600; white-space: nowrap; }
        .tag-error { background: #fee2e2; color: #b91c1c; }
        .tag-limit { background: #fef3c7; color: #b45309; }
        .tag-temp { background: #e0e7ff; color: #4338ca; }
        .tag-ok { background: #d1fae5; color: #047857; }
        .low { color: #dc2626; font-weight: bold; }
        .ok { color: #059669; font-weight: bold; }
        .group-block { margin-bottom: 26px; }
        .info { color: #666; font-size: 13px; line-height: 1.7; margin-top: 20px; text-align: center; }
        .btn-mail { display: inline-block; margin-top: 14px; padding: 10px 26px; color: #fff; text-decoration: none; border-radius: 6px; font-size: 14px; font-weight: bold; }
        .footer { background-color: #f8f9fa; padding: 16px; text-align: center; color: #999; font-size: 12px; }
`

func renderAccountAlertEmail(siteName, headerClass, title, content string) string {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html>\n<head>\n    <meta charset=\"UTF-8\">\n    <style>")
	b.WriteString(accountAlertEmailStyle)
	b.WriteString("    </style>\n</head>\n<body>\n    <div class=\"container\">\n")
	b.WriteString(`        <div class="header ` + headerClass + `"><h1>` + html.EscapeString(title) + `</h1><p>` + html.EscapeString(siteName) + `</p></div>` + "\n")
	b.WriteString(`        <div class="content">` + "\n")
	b.WriteString(content)
	b.WriteString("\n        </div>\n")
	b.WriteString(`        <div class="footer"><p>此邮件由系统自动发送，请勿回复。</p></div>` + "\n")
	b.WriteString("    </div>\n</body>\n</html>")
	return b.String()
}

func accountAlertMetric(label, value, valueClass string) string {
	class := "metric-value"
	if valueClass != "" {
		class += " " + valueClass
	}
	return `<div class="metric"><span class="metric-label">` + html.EscapeString(label) + `</span><span class="` + class + `">` + html.EscapeString(value) + `</span></div>`
}

func accountAlertAccountLabel(id int64, name string) string {
	return "#" + strconv.FormatInt(id, 10) + " " + name
}

func accountAlertManageButton(adminURL, headerClass string) string {
	if adminURL == "" {
		return ""
	}
	return `<a class="btn-mail ` + headerClass + `" href="` + html.EscapeString(adminURL+"/admin/accounts") + `">前往账号管理</a>`
}

func buildAccountErrorAlertEmail(siteName, adminURL string, rows []accountAlertErrorRow, detectedAt time.Time) (string, string) {
	subject := fmt.Sprintf("[%s] %d 个账号变为错误状态", siteName, len(rows))
	if len(rows) == 1 {
		subject = fmt.Sprintf("[%s] 账号「%s」变为错误状态", siteName, rows[0].Name)
	}

	var b strings.Builder
	b.WriteString(`<p class="lead">以下账号从正常变为错误状态，已停止调度</p>`)
	b.WriteString(`<table><thead><tr><th>账号</th><th>平台</th><th>所属分组</th></tr></thead><tbody>`)
	for _, row := range rows {
		groups := "-"
		if len(row.Groups) > 0 {
			groups = strings.Join(row.Groups, "、")
		}
		b.WriteString(`<tr><td><b>` + html.EscapeString(accountAlertAccountLabel(row.ID, row.Name)) + `</b></td><td>` +
			html.EscapeString(row.Platform+" / "+row.Type) + `</td><td>` + html.EscapeString(groups) + `</td></tr>`)
		reason := row.Reason
		if reason == "" {
			reason = "-"
		}
		b.WriteString(`<tr><td colspan="3" class="reason">原因：` + html.EscapeString(reason) + `</td></tr>`)
	}
	b.WriteString(`</tbody></table>`)
	b.WriteString(`<div class="info"><p>发现时间：` + html.EscapeString(detectedAt.Format(accountAlertTimeLayout)) + `</p><p>处理后可在「账号管理」中恢复账号状态。</p>` +
		accountAlertManageButton(adminURL, "red") + `</div>`)
	return subject, renderAccountAlertEmail(siteName, "red", "账号异常告警", b.String())
}

func buildGroupUnavailableAlertEmail(siteName, adminURL string, groups []accountAlertGroupDownItem, now time.Time) (string, string) {
	subject := fmt.Sprintf("[%s] %d 个分组无可用账号", siteName, len(groups))
	if len(groups) == 1 {
		subject = fmt.Sprintf("[%s] 分组「%s」无可用账号", siteName, groups[0].Name)
	}

	var b strings.Builder
	b.WriteString(`<p class="lead">以下分组当前没有可用账号，用户请求会失败</p>`)
	for _, group := range groups {
		b.WriteString(`<div class="group-block">`)
		b.WriteString(accountAlertMetric("分组", group.Name+"（#"+strconv.FormatInt(group.ID, 10)+"）", ""))
		b.WriteString(accountAlertMetric("平台", group.Platform, ""))
		b.WriteString(accountAlertMetric("可用账号", "0 / "+strconv.Itoa(group.Total), "low"))
		b.WriteString(accountAlertMetric("开始不可用", group.ZeroSince.Format(accountAlertTimeLayout)+"（已持续 "+formatAccountAlertDuration(now.Sub(group.ZeroSince))+"）", ""))
		b.WriteString(`<h4>组内账号</h4><table><thead><tr><th>账号</th><th>状态</th><th>原因 / 预计恢复</th></tr></thead><tbody>`)
		for _, member := range group.Members {
			detail := ""
			if member.Detail != "" {
				detail = `<span class="reason">` + html.EscapeString(member.Detail) + `</span><br>`
			}
			recovery := ""
			if member.Recovery != "" {
				recovery = `<span class="muted">` + html.EscapeString(member.Recovery) + `</span>`
			}
			if detail == "" && recovery == "" {
				recovery = "-"
			}
			b.WriteString(`<tr><td><b>` + html.EscapeString(accountAlertAccountLabel(member.ID, member.Name)) + `</b></td><td><span class="tag ` +
				member.TagClass + `">` + html.EscapeString(member.Label) + `</span></td><td>` + detail + recovery + `</td></tr>`)
		}
		b.WriteString(`</tbody></table></div>`)
	}
	b.WriteString(`<div class="info">` + accountAlertManageButton(adminURL, "red") + `</div>`)
	return subject, renderAccountAlertEmail(siteName, "red", "分组不可用告警", b.String())
}

func buildGroupRecoveredAlertEmail(siteName string, groups []accountAlertGroupUpItem) (string, string) {
	subject := fmt.Sprintf("[%s] %d 个分组已恢复", siteName, len(groups))
	if len(groups) == 1 {
		subject = fmt.Sprintf("[%s] 分组「%s」已恢复", siteName, groups[0].Name)
	}

	var b strings.Builder
	b.WriteString(`<p class="lead">以下分组已重新有可用账号</p>`)
	for _, group := range groups {
		b.WriteString(`<div class="group-block">`)
		b.WriteString(accountAlertMetric("分组", group.Name+"（#"+strconv.FormatInt(group.ID, 10)+"）", ""))
		b.WriteString(accountAlertMetric("可用账号", strconv.Itoa(group.Available)+" / "+strconv.Itoa(group.Total), "ok"))
		b.WriteString(accountAlertMetric("不可用时段", group.ZeroSince.Format(accountAlertTimeLayout)+" → "+group.RecoveredAt.Format(accountAlertTimeLayout), ""))
		b.WriteString(accountAlertMetric("不可用时长", formatAccountAlertDuration(group.RecoveredAt.Sub(group.ZeroSince)), ""))
		if len(group.Members) > 0 {
			b.WriteString(`<h4>当前可用账号</h4><table><tbody>`)
			for _, name := range group.Members {
				b.WriteString(`<tr><td><b>` + html.EscapeString(name) + `</b></td><td><span class="tag tag-ok">可用</span></td></tr>`)
			}
			b.WriteString(`</tbody></table>`)
		}
		b.WriteString(`</div>`)
	}
	return subject, renderAccountAlertEmail(siteName, "green", "分组已恢复", b.String())
}

func buildUpstreamBalanceLowAlertEmail(siteName string, threshold float64, hosts []accountAlertBalanceHostItem) (string, string) {
	thresholdText := formatAccountAlertAmount(threshold)
	subject := fmt.Sprintf("[%s] %d 个上游余额不足", siteName, len(hosts))
	if len(hosts) == 1 {
		subject = fmt.Sprintf("[%s] 上游 %s 余额不足（%s < %s）", siteName, hosts[0].Host, formatAccountAlertAmount(hosts[0].MinBalance), thresholdText)
	}

	var b strings.Builder
	b.WriteString(`<p class="lead">以下上游的余额已低于提醒阈值 <b>` + html.EscapeString(thresholdText) + `</b></p>`)
	for _, host := range hosts {
		b.WriteString(`<div class="group-block">`)
		b.WriteString(accountAlertMetric("上游", host.Host, ""))
		b.WriteString(accountAlertMetric("最低余额", formatAccountAlertBalance(host.MinBalance, host.Unit), "low"))
		b.WriteString(accountAlertMetric("提醒阈值", thresholdText, ""))
		b.WriteString(`<h4>该上游下的账号</h4><table><thead><tr><th>账号</th><th>余额</th><th>查询时间</th></tr></thead><tbody>`)
		for _, acc := range host.Accounts {
			balanceClass := ""
			if acc.Low {
				balanceClass = ` class="low"`
			}
			receivedAt := "-"
			if acc.ReceivedAt != nil {
				receivedAt = acc.ReceivedAt.Local().Format(accountAlertTimeLayout)
			}
			b.WriteString(`<tr><td><b>` + html.EscapeString(accountAlertAccountLabel(acc.ID, acc.Name)) + `</b></td><td` + balanceClass + `>` +
				html.EscapeString(formatAccountAlertBalance(acc.Balance, acc.Unit)) + `</td><td>` + html.EscapeString(receivedAt) + `</td></tr>`)
		}
		b.WriteString(`</tbody></table></div>`)
	}
	b.WriteString(`<div class="info"><p>请及时给上游充值，以免账号因余额不足被停用。</p><p class="muted">余额回升到阈值以上后，再次跌破才会重新通知。</p></div>`)
	return subject, renderAccountAlertEmail(siteName, "amber", "上游余额不足", b.String())
}

// BuildAccountAlertNotifyTestEmail 生成「发送测试通知」使用的示例邮件。
func BuildAccountAlertNotifyTestEmail(siteName string) (string, string) {
	subject := fmt.Sprintf("[%s] 账号与上游监控通知测试", siteName)
	var b strings.Builder
	b.WriteString(`<p class="lead">收到这封邮件，说明监控通知可以正常发送</p>`)
	b.WriteString(accountAlertMetric("账号出错通知", "账号从正常变为错误时发送", ""))
	b.WriteString(accountAlertMetric("分组无可用账号通知", "分组可用账号数降为 0 时发送", ""))
	b.WriteString(accountAlertMetric("上游余额不足通知", "上游余额低于阈值时发送", ""))
	b.WriteString(`<div class="info"><p>发送时间：` + html.EscapeString(time.Now().Format(accountAlertTimeLayout)) + `</p></div>`)
	return subject, renderAccountAlertEmail(siteName, "blue", "监控通知测试", b.String())
}

func formatAccountAlertDuration(d time.Duration) string {
	if d < time.Minute {
		return "不到 1 分钟"
	}
	totalMinutes := int64(d / time.Minute)
	days := totalMinutes / (24 * 60)
	hours := (totalMinutes % (24 * 60)) / 60
	minutes := totalMinutes % 60
	parts := make([]string, 0, 3)
	if days > 0 {
		parts = append(parts, strconv.FormatInt(days, 10)+" 天")
	}
	if hours > 0 {
		parts = append(parts, strconv.FormatInt(hours, 10)+" 小时")
	}
	if minutes > 0 || len(parts) == 0 {
		parts = append(parts, strconv.FormatInt(minutes, 10)+" 分钟")
	}
	return strings.Join(parts, " ")
}

func formatAccountAlertAmount(v float64) string {
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64)
}

func formatAccountAlertBalance(v float64, unit string) string {
	text := strconv.FormatFloat(v, 'f', 2, 64)
	if unit = strings.TrimSpace(unit); unit != "" {
		text += " " + unit
	}
	return text
}

// truncateAccountAlertText 按字符截断，避免超长的上游错误信息撑大邮件。
func truncateAccountAlertText(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "…"
}
