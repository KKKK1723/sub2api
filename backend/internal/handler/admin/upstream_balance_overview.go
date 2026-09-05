package admin

import (
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type upstreamBalanceOverviewItem struct {
	BaseURL      string     `json:"base_url"`
	AccountID    int64      `json:"account_id"`
	GroupIDs     []int64    `json:"group_ids,omitempty"`
	GroupNames   []string   `json:"group_names,omitempty"`
	Balance      *float64   `json:"balance,omitempty"`
	Unit         string     `json:"unit,omitempty"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
	Status       string     `json:"status"`
}

type upstreamBalanceRepresentative struct {
	baseURL   string
	accountID int64
	groupIDs  []int64
	groupNames []string
}

func upstreamBalanceRepresentatives(accounts []service.Account) []upstreamBalanceRepresentative {
	byBaseURL := make(map[string]int64)
	byGroups := make(map[string]struct{ groupIDs []int64; groupNames []string })
	for _, account := range accounts {
		if account.Type != service.AccountTypeAPIKey || !account.IsActive() || account.GetCredential("api_key") == "" {
			continue
		}
		baseURL := canonicalOverviewBaseURL(service.UpstreamBalanceBaseURL(&account))
		if baseURL == "" {
			continue
		}
		if _, exists := byBaseURL[baseURL]; !exists {
			byBaseURL[baseURL] = account.ID
		}
		groups := byGroups[baseURL]
		seen := make(map[int64]struct{}, len(groups.groupIDs))
		for i, gid := range groups.groupIDs { seen[gid] = struct{}{}; _ = i }
		for _, ag := range account.AccountGroups {
			if _, ok := seen[ag.GroupID]; ok { continue }
			seen[ag.GroupID] = struct{}{}
			groups.groupIDs = append(groups.groupIDs, ag.GroupID)
			if ag.Group != nil { groups.groupNames = append(groups.groupNames, ag.Group.Name) }
		}
		byGroups[baseURL] = groups
	}
	items := make([]upstreamBalanceRepresentative, 0, len(byBaseURL))
	for baseURL, accountID := range byBaseURL {
		groups := byGroups[baseURL]
		items = append(items, upstreamBalanceRepresentative{baseURL: baseURL, accountID: accountID, groupIDs: groups.groupIDs, groupNames: groups.groupNames})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].baseURL < items[j].baseURL })
	return items
}

func canonicalOverviewBaseURL(raw string) string {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" { return raw }
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	u.Path = strings.TrimRight(u.Path, "/")
	u.RawQuery, u.Fragment = "", ""
	return strings.TrimRight(u.String(), "/")
}

// RefreshUpstreamBalanceOverview refreshes one representative account per site.
// The account's API key stays inside the probe service; only public site and
// balance metadata are returned to the administrator.
func (h *AccountHandler) RefreshUpstreamBalanceOverview(c *gin.Context) {
	if h.upstreamBillingProbe == nil {
		response.ErrorFrom(c, service.ErrUpstreamBalanceProbeUnavailable)
		return
	}
	accounts, _, err := h.adminService.ListAccounts(
		c.Request.Context(), 1, 10000, "", service.AccountTypeAPIKey, service.StatusActive, "", 0, "", "id", "asc",
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	representatives := upstreamBalanceRepresentatives(accounts)
	items := make([]upstreamBalanceOverviewItem, len(representatives))
	var itemsMu sync.Mutex
	group, ctx := errgroup.WithContext(c.Request.Context())
	group.SetLimit(10)
	for index, representative := range representatives {
		index, representative := index, representative
		items[index] = upstreamBalanceOverviewItem{BaseURL: representative.baseURL, AccountID: representative.accountID, GroupIDs: representative.groupIDs, GroupNames: representative.groupNames, Status: service.UpstreamBalanceProbeStatusFailed}
		group.Go(func() error {
			snapshot, probeErr := h.upstreamBillingProbe.ProbeUpstreamBalance(ctx, representative.accountID)
			if probeErr != nil || snapshot == nil {
				return nil
			}
			item := upstreamBalanceOverviewItem{
				BaseURL: representative.baseURL,
				AccountID: representative.accountID,
				GroupIDs: representative.groupIDs,
				GroupNames: representative.groupNames,
				Status:  snapshot.Status,
			}
			if !snapshot.LastAttemptAt.IsZero() {
				attemptedAt := snapshot.LastAttemptAt
				item.LastSyncedAt = &attemptedAt
			}
			if snapshot.Data != nil {
				balance := snapshot.Data.Balance
				item.Balance = &balance
				item.Unit = snapshot.Data.Unit
			}
			itemsMu.Lock()
			items[index] = item
			itemsMu.Unlock()
			return nil
		})
	}
	_ = group.Wait()
	response.Success(c, items)
}

type upstreamBalanceWriteRequest struct {
	BaseURL string `json:"base_url" binding:"required"`
	APIKey string `json:"api_key"`
	Name string `json:"name"`
	Platform string `json:"platform"`
	GroupIDs []int64 `json:"group_ids"`
}

func (h *AccountHandler) CreateUpstreamBalance(c *gin.Context) {
	var req upstreamBalanceWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "invalid request: "+err.Error()); return }
	base := canonicalOverviewBaseURL(req.BaseURL)
	if base == "" || req.APIKey == "" { response.BadRequest(c, "base_url and api_key are required"); return }
	platform := req.Platform; if platform == "" { platform = service.PlatformOpenAI }
	name := strings.TrimSpace(req.Name); if name == "" { name = base }
	account, err := h.adminService.CreateAccount(c.Request.Context(), &service.CreateAccountInput{
		Name: name, Platform: platform, Type: service.AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": req.APIKey, "base_url": base}, GroupIDs: req.GroupIDs,
		Extra: map[string]any{service.UpstreamBalanceProbeEnabledExtraKey: true, service.UpstreamBalanceQueryExtraKey: map[string]any{"preset": service.UpstreamBalancePresetCCSwitch}},
		SkipDefaultGroupBind: true, SkipMixedChannelCheck: true,
	})
	if err != nil { response.ErrorFrom(c, err); return }
	response.Success(c, h.accountResponseFromService(account))
}

func (h *AccountHandler) UpdateUpstreamBalance(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64); if err != nil { response.BadRequest(c, "invalid account id"); return }
	var req upstreamBalanceWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.BadRequest(c, "invalid request: "+err.Error()); return }
	account, err := h.adminService.GetAccount(c.Request.Context(), id); if err != nil { response.ErrorFrom(c, err); return }
	credentials := map[string]any{}
	for k, v := range account.Credentials { credentials[k] = v }
	if req.APIKey != "" { credentials["api_key"] = req.APIKey }
	if req.BaseURL != "" { credentials["base_url"] = canonicalOverviewBaseURL(req.BaseURL) }
	name := req.Name; if strings.TrimSpace(name) == "" { name = account.Name }
	updated, err := h.adminService.UpdateAccount(c.Request.Context(), id, &service.UpdateAccountInput{Name: name, Credentials: credentials, GroupIDs: &req.GroupIDs, SkipMixedChannelCheck: true})
	if err != nil { response.ErrorFrom(c, err); return }
	response.Success(c, h.accountResponseFromService(updated))
}

func (h *AccountHandler) DeleteUpstreamBalance(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64); if err != nil { response.BadRequest(c, "invalid account id"); return }
	if _, err := h.adminService.UpdateAccount(c.Request.Context(), id, &service.UpdateAccountInput{Status: "inactive"}); err != nil { response.ErrorFrom(c, err); return }
	response.Success(c, gin.H{"message": "upstream balance disabled"})
}
