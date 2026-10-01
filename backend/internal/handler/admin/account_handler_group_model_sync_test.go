package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type groupModelSyncAdmin struct {
	*stubAdminService
	group service.Group
}

func (s *groupModelSyncAdmin) GetGroup(context.Context, int64) (*service.Group, error) {
	group := s.group
	return &group, nil
}

func (s *groupModelSyncAdmin) UpdateGroup(_ context.Context, _ int64, input *service.UpdateGroupInput) (*service.Group, error) {
	if input.ModelsListConfig != nil {
		s.group.ModelsListConfig = *input.ModelsListConfig
	}
	return &s.group, nil
}

type groupModelSyncUpstream struct {
	models []string
	fail   bool
	calls  int
}

func (u *groupModelSyncUpstream) Do(*http.Request, string, int64, int) (*http.Response, error) {
	u.calls++
	if u.fail {
		return nil, fmt.Errorf("private upstream failure")
	}
	data := make([]map[string]string, 0, len(u.models))
	for _, model := range u.models {
		data = append(data, map[string]string{"id": model})
	}
	body, _ := json.Marshal(map[string]any{"data": data})
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(string(body)))}, nil
}

func (u *groupModelSyncUpstream) DoWithTLS(req *http.Request, proxy string, id int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxy, id, concurrency)
}

func groupModelAccount(id int64, mapping map[string]any) service.Account {
	credentials := map[string]any{"api_key": "test-key", "base_url": "https://api.openai.com"}
	if mapping != nil {
		credentials["model_mapping"] = mapping
	}
	return service.Account{ID: id, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey,
		Status: service.StatusActive, Schedulable: true, Credentials: credentials}
}

func runGroupModelSync(t *testing.T, accounts []service.Account, upstream *groupModelSyncUpstream) (*httptest.ResponseRecorder, *groupModelSyncAdmin) {
	t.Helper()
	svc := &groupModelSyncAdmin{stubAdminService: newStubAdminService(), group: service.Group{
		ID: 3, Name: "GPT-PRO", Platform: service.PlatformOpenAI,
		ModelsListConfig: service.GroupModelsListConfig{Enabled: true, Models: []string{"old-model"}},
	}}
	svc.accounts = accounts
	testSvc := service.NewAccountTestService(nil, nil, nil, nil, nil, upstream,
		&config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}}, nil)
	handler := NewAccountHandler(svc, nil, nil, nil, nil, nil, nil, nil, testSvc, nil, nil, nil, nil, nil)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/groups/:group_id/sync", handler.SyncAvailableModels)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/groups/3/sync", nil))
	return rec, svc
}

func TestSyncAvailableModelsRespectsAllFourAccountWhitelists(t *testing.T) {
	models := []string{"codex-auto-review", "gpt-5.5", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-6-astra", "gpt-6-sol"}
	accounts := make([]service.Account, 0, 4)
	for _, id := range []int64{53, 72, 75, 77} {
		mapping := make(map[string]any)
		for _, model := range models {
			mapping[model] = model
		}
		accounts = append(accounts, groupModelAccount(id, mapping))
	}
	upstream := &groupModelSyncUpstream{models: append(append([]string{}, models...), "gpt-image-2", "gpt-image-2.5-flare")}
	rec, svc := runGroupModelSync(t, accounts, upstream)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, models, svc.group.ModelsListConfig.Models)
	require.Zero(t, upstream.calls, "明确白名单不应被上游完整目录覆盖")
	require.NotNil(t, svc.group.ModelsListConfig.LastSyncedAt)
}

func TestSyncAvailableModelsAggregatesEveryPageAndKeepsRequestAliases(t *testing.T) {
	accounts := make([]service.Account, 0, 101)
	for i := int64(1); i <= 101; i++ {
		accounts = append(accounts, groupModelAccount(i, map[string]any{"shared": "provider-private"}))
	}
	accounts[100].Credentials["model_mapping"] = map[string]any{"last-page-alias": "provider-last"}
	rec, svc := runGroupModelSync(t, accounts, &groupModelSyncUpstream{fail: true})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, []string{"last-page-alias", "shared"}, svc.group.ModelsListConfig.Models)
	require.Equal(t, 2, svc.lastListAccounts.calls)
	require.Equal(t, int64(3), svc.lastListAccounts.groupID)
}

func TestSyncAvailableModelsExpandsOnlyAllowedWildcards(t *testing.T) {
	accounts := []service.Account{groupModelAccount(1, map[string]any{"gpt-5.*": "upstream-fixed", "public-alias": "provider-private"})}
	upstream := &groupModelSyncUpstream{models: []string{"gpt-5.5", "gpt-5.6-sol", "gpt-image-2", "provider-private", "gpt-5.5"}}
	rec, svc := runGroupModelSync(t, accounts, upstream)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, []string{"gpt-5.5", "gpt-5.6-sol", "public-alias"}, svc.group.ModelsListConfig.Models)
	require.Equal(t, 1, upstream.calls)
}

func TestSyncAvailableModelsUnrestrictedAndPassthrough(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		t.Run(fmt.Sprint(passthrough), func(t *testing.T) {
			account := groupModelAccount(1, nil)
			if passthrough {
				account.Credentials["model_mapping"] = map[string]any{"stale-model": "stale-model"}
				account.Extra = map[string]any{"openai_passthrough": true}
			}
			rec, svc := runGroupModelSync(t, []service.Account{account}, &groupModelSyncUpstream{models: []string{"upstream-model"}})
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			require.Equal(t, []string{"upstream-model"}, svc.group.ModelsListConfig.Models)
		})
	}
}

func TestSyncAvailableModelsSkipsDisabledAccountsButKeepsTemporaryLimits(t *testing.T) {
	accounts := make([]service.Account, 4)
	for i := range accounts {
		model := fmt.Sprintf("model-%d", i)
		accounts[i] = groupModelAccount(int64(i+1), map[string]any{model: model})
	}
	accounts[0].Status = "inactive"
	accounts[1].Schedulable = false
	expired := time.Now().Add(-time.Hour)
	accounts[2].ExpiresAt, accounts[2].AutoPauseOnExpired = &expired, true
	reset := time.Now().Add(time.Hour)
	accounts[3].RateLimitResetAt = &reset
	rec, svc := runGroupModelSync(t, accounts, &groupModelSyncUpstream{})
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, []string{"model-3"}, svc.group.ModelsListConfig.Models)
}

func TestSyncAvailableModelsPreservesPreviousListOnIncompleteOrEmptyResult(t *testing.T) {
	for _, test := range []struct {
		name     string
		accounts []service.Account
		upstream *groupModelSyncUpstream
	}{
		{"partial-failure", []service.Account{groupModelAccount(1, map[string]any{"new-model": "new-model"}), groupModelAccount(2, nil)}, &groupModelSyncUpstream{fail: true}},
		{"empty-group", nil, &groupModelSyncUpstream{}},
		{"filtered-empty", []service.Account{groupModelAccount(1, map[string]any{"gpt-*": "gpt"})}, &groupModelSyncUpstream{models: []string{"claude-model"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			rec, svc := runGroupModelSync(t, test.accounts, test.upstream)
			require.Equal(t, http.StatusBadGateway, rec.Code)
			require.Equal(t, []string{"old-model"}, svc.group.ModelsListConfig.Models)
			require.NotEmpty(t, svc.group.ModelsListConfig.SyncError)
			require.NotContains(t, rec.Body.String(), "private upstream failure")
		})
	}
}
