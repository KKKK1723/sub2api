package admin

import (
	"log/slog"
	"net/mail"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// AccountAlertNotifySettingsResponse 账号与上游监控通知配置
type AccountAlertNotifySettingsResponse struct {
	Emails                       []dto.NotifyEmailEntry `json:"emails"`
	AccountErrorEnabled          bool                   `json:"account_error_enabled"`
	GroupUnavailableEnabled      bool                   `json:"group_unavailable_enabled"`
	GroupUnavailableDelayMinutes int                    `json:"group_unavailable_delay_minutes"`
	GroupRecoveredEnabled        bool                   `json:"group_recovered_enabled"`
	UpstreamBalanceLowEnabled    bool                   `json:"upstream_balance_low_enabled"`
	UpstreamBalanceLowThreshold  float64                `json:"upstream_balance_low_threshold"`
}

func accountAlertNotifySettingsResponse(cfg *service.AccountAlertNotifyConfig) AccountAlertNotifySettingsResponse {
	emails := dto.NotifyEmailEntriesFromService(cfg.Emails)
	if emails == nil {
		emails = []dto.NotifyEmailEntry{}
	}
	return AccountAlertNotifySettingsResponse{
		Emails:                       emails,
		AccountErrorEnabled:          cfg.AccountErrorEnabled,
		GroupUnavailableEnabled:      cfg.GroupUnavailableEnabled,
		GroupUnavailableDelayMinutes: cfg.GroupUnavailableDelayMinutes,
		GroupRecoveredEnabled:        cfg.GroupRecoveredEnabled,
		UpstreamBalanceLowEnabled:    cfg.UpstreamBalanceLowEnabled,
		UpstreamBalanceLowThreshold:  cfg.UpstreamBalanceLowThreshold,
	}
}

// GetAccountAlertNotifySettings 获取账号与上游监控通知配置
// GET /api/v1/admin/settings/account-alert-notify
func (h *SettingHandler) GetAccountAlertNotifySettings(c *gin.Context) {
	cfg, err := h.settingService.GetAccountAlertNotifyConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, accountAlertNotifySettingsResponse(cfg))
}

// UpdateAccountAlertNotifySettingsRequest 更新账号与上游监控通知配置请求
type UpdateAccountAlertNotifySettingsRequest struct {
	Emails                       []dto.NotifyEmailEntry `json:"emails"`
	AccountErrorEnabled          bool                   `json:"account_error_enabled"`
	GroupUnavailableEnabled      bool                   `json:"group_unavailable_enabled"`
	GroupUnavailableDelayMinutes int                    `json:"group_unavailable_delay_minutes"`
	GroupRecoveredEnabled        bool                   `json:"group_recovered_enabled"`
	UpstreamBalanceLowEnabled    bool                   `json:"upstream_balance_low_enabled"`
	UpstreamBalanceLowThreshold  float64                `json:"upstream_balance_low_threshold"`
}

// UpdateAccountAlertNotifySettings 更新账号与上游监控通知配置
// PUT /api/v1/admin/settings/account-alert-notify
func (h *SettingHandler) UpdateAccountAlertNotifySettings(c *gin.Context) {
	var req UpdateAccountAlertNotifySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	emails := dto.NotifyEmailEntriesToService(req.Emails)
	if emails == nil {
		emails = []service.NotifyEmailEntry{}
	}
	cfg := &service.AccountAlertNotifyConfig{
		Emails:                       emails,
		AccountErrorEnabled:          req.AccountErrorEnabled,
		GroupUnavailableEnabled:      req.GroupUnavailableEnabled,
		GroupUnavailableDelayMinutes: req.GroupUnavailableDelayMinutes,
		GroupRecoveredEnabled:        req.GroupRecoveredEnabled,
		UpstreamBalanceLowEnabled:    req.UpstreamBalanceLowEnabled,
		UpstreamBalanceLowThreshold:  req.UpstreamBalanceLowThreshold,
	}
	if err := h.settingService.SetAccountAlertNotifyConfig(c.Request.Context(), cfg); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	updated, err := h.settingService.GetAccountAlertNotifyConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, accountAlertNotifySettingsResponse(updated))
}

// SendAccountAlertNotifyTestRequest 发送测试通知请求；未传邮箱时使用已保存的收件邮箱。
type SendAccountAlertNotifyTestRequest struct {
	Emails []string `json:"emails"`
}

// SendAccountAlertNotifyTest 发送账号与上游监控通知测试邮件
// POST /api/v1/admin/settings/account-alert-notify/test
func (h *SettingHandler) SendAccountAlertNotifyTest(c *gin.Context) {
	var req SendAccountAlertNotifyTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	ctx := c.Request.Context()
	recipients := make([]string, 0, len(req.Emails))
	seen := make(map[string]struct{}, len(req.Emails))
	for _, email := range req.Emails {
		email = strings.TrimSpace(email)
		if email == "" {
			continue
		}
		if _, err := mail.ParseAddress(email); err != nil || strings.ContainsAny(email, "<> ") {
			response.BadRequest(c, "Invalid email: "+email)
			return
		}
		lower := strings.ToLower(email)
		if _, ok := seen[lower]; ok {
			continue
		}
		seen[lower] = struct{}{}
		recipients = append(recipients, email)
	}
	if len(recipients) == 0 {
		cfg, err := h.settingService.GetAccountAlertNotifyConfig(ctx)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		recipients = cfg.Recipients()
	}
	if len(recipients) == 0 {
		response.BadRequest(c, "No recipient email configured")
		return
	}
	// 与保存时的收件人上限一致，防止被用来批量发信
	if len(recipients) > 10 {
		response.BadRequest(c, "At most 10 emails are allowed")
		return
	}

	subject, body := service.BuildAccountAlertNotifyTestEmail(h.settingService.GetSiteName(ctx))
	sent := 0
	var lastErr error
	for _, to := range recipients {
		if err := h.emailService.SendEmail(ctx, to, subject, body); err != nil {
			slog.Warn("account_alert_notify.test_send_failed", "to", to, "error", err)
			lastErr = err
			continue
		}
		sent++
	}
	if sent == 0 {
		msg := "Failed to send test email"
		if lastErr != nil {
			msg += ": " + lastErr.Error()
		}
		response.BadRequest(c, msg)
		return
	}

	response.Success(c, gin.H{
		"sent":  sent,
		"total": len(recipients),
	})
}
