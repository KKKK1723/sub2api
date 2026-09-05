package service

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/stretchr/testify/require"
)

func TestNormalizeUpstreamBalanceQueryCustom(t *testing.T) {
	query, err := NormalizeUpstreamBalanceQuery(&UpstreamBalanceQuery{
		Preset:      " custom ",
		Method:      "post",
		URL:         "{{origin}}/api/usage",
		AuthType:    "x-api-key",
		BalancePath: "response.balance",
		UnitPath:    "response.unit",
		ErrorPath:   "response.error",
		Unit:        "usd",
	})
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, query.Method)
	require.Equal(t, "USD", query.Unit)
	require.Equal(t, "response.balance", query.BalancePath)

	_, err = NormalizeUpstreamBalanceQuery(&UpstreamBalanceQuery{
		Preset: UpstreamBalancePresetCustom, URL: "https://other.example/usage", BalancePath: "response[0].balance",
	})
	require.Error(t, err)

	_, err = NormalizeUpstreamBalanceQuery(&UpstreamBalanceQuery{
		Preset: UpstreamBalancePresetCustom, URL: "{{baseUrl}}/usage?key={{apiKey}}", BalancePath: "balance",
	})
	require.Error(t, err)
}

func TestParseUpstreamBalanceResponsePresets(t *testing.T) {
	sub2api, err := resolveUpstreamBalanceQuery(&UpstreamBalanceQuery{Preset: UpstreamBalancePresetSub2API})
	require.NoError(t, err)
	data, err := parseUpstreamBalanceResponse([]byte(`{"quota":{"remaining":19.5,"unit":"USD","limit":25,"used":5.5},"planName":"Team","isValid":true}`), sub2api)
	require.NoError(t, err)
	require.Equal(t, 19.5, data.Balance)
	require.Equal(t, "USD", data.Unit)
	require.Equal(t, "Team", data.PlanName)
	require.Equal(t, 25.0, *data.Total)
	require.Equal(t, 5.5, *data.Used)
	require.True(t, *data.IsValid)

	ccswitch, err := resolveUpstreamBalanceQuery(&UpstreamBalanceQuery{Preset: UpstreamBalancePresetCCSwitch})
	require.NoError(t, err)
	require.Equal(t, sub2api, ccswitch)
	data, err = parseUpstreamBalanceResponse([]byte(`{"remaining":"8.25","unit":"CNY","is_active":true}`), ccswitch)
	require.NoError(t, err)
	require.Equal(t, 8.25, data.Balance)
	require.Equal(t, "CNY", data.Unit)
	require.True(t, *data.IsValid)

	_, err = parseUpstreamBalanceResponse([]byte(`{"error":"invalid key","balance":8.25}`), ccswitch)
	require.Error(t, err)
}

func TestResolveUpstreamBalanceURL(t *testing.T) {
	value, err := resolveUpstreamBalanceURL("https://example.com/v1", "{{baseUrl}}/v1/usage")
	require.NoError(t, err)
	require.Equal(t, "https://example.com/v1/usage", value)

	value, err = resolveUpstreamBalanceURL("https://example.com/v1", "{{origin}}/api/usage")
	require.NoError(t, err)
	require.Equal(t, "https://example.com/api/usage", value)
	require.True(t, sameUpstreamBalanceOrigin("https://example.com/v1", value))
	require.False(t, sameUpstreamBalanceOrigin("https://example.com/v1", "https://other.example/api/usage"))
}

func TestUpstreamBalanceBaseURLCanonicalizesSiteAddress(t *testing.T) {
	account := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"base_url": "https://balance.example/api///?ignored=yes",
		},
	}
	require.Equal(t, "https://balance.example/api", UpstreamBalanceBaseURL(account))

	account.Credentials = map[string]any{}
	require.Equal(t, "https://api.openai.com", UpstreamBalanceBaseURL(account))
}

func TestUpstreamBalanceQueryForAccountDefaultsToCCSwitch(t *testing.T) {
	query := upstreamBalanceQueryForAccount(&Account{Extra: map[string]any{}})
	require.NotNil(t, query)
	require.Equal(t, UpstreamBalancePresetCCSwitch, query.Preset)
}

type upstreamBalanceAccountRepo struct {
	*upstreamBillingProbeAccountRepo
}

func (r *upstreamBalanceAccountRepo) UpdateUpstreamBalanceProbeSnapshot(_ context.Context, expected *Account, snapshot *UpstreamBalanceProbeSnapshot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	account := r.accounts[expected.ID]
	if account == nil || account.Platform != expected.Platform || account.Type != expected.Type ||
		!reflect.DeepEqual(account.Credentials, expected.Credentials) ||
		!reflect.DeepEqual(account.Extra[UpstreamBalanceQueryExtraKey], expected.Extra[UpstreamBalanceQueryExtraKey]) {
		return ErrUpstreamBalanceProbeIdentityChanged
	}
	if account.Extra == nil {
		account.Extra = make(map[string]any)
	}
	account.Extra[UpstreamBalanceProbeExtraKey] = snapshot
	return nil
}

type upstreamBalanceHTTPStub struct {
	mu      sync.Mutex
	request *http.Request
}

func (u *upstreamBalanceHTTPStub) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	u.mu.Lock()
	u.request = req.Clone(req.Context())
	u.mu.Unlock()
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"balance":"12.34","unit":"USD","planName":"Pro"}`)),
	}, nil
}

func (u *upstreamBalanceHTTPStub) DoWithTLS(req *http.Request, proxyURL string, accountID int64, concurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxyURL, accountID, concurrency)
}

func TestProbeUpstreamBalanceCCSwitch(t *testing.T) {
	baseRepo := &upstreamBillingProbeAccountRepo{accounts: map[int64]*Account{
		7: {
			ID: 7, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive,
			Credentials: map[string]any{"base_url": "https://example.com/v1", "api_key": "secret"},
			Extra: map[string]any{
				UpstreamBalanceProbeEnabledExtraKey: true,
				UpstreamBalanceQueryExtraKey:        &UpstreamBalanceQuery{Preset: UpstreamBalancePresetCCSwitch},
			},
		},
	}}
	repo := &upstreamBalanceAccountRepo{upstreamBillingProbeAccountRepo: baseRepo}
	httpStub := &upstreamBalanceHTTPStub{}
	svc := newUpstreamBillingProbeTestService(repo, httpStub, nil)

	snapshot, err := svc.ProbeUpstreamBalance(context.Background(), 7)
	require.NoError(t, err)
	require.Equal(t, UpstreamBalanceProbeStatusOK, snapshot.Status)
	require.Equal(t, 12.34, snapshot.Data.Balance)
	require.Equal(t, "Pro", snapshot.Data.PlanName)

	httpStub.mu.Lock()
	request := httpStub.request
	httpStub.mu.Unlock()
	require.NotNil(t, request)
	require.Equal(t, http.MethodGet, request.Method)
	require.Equal(t, "https://example.com/v1/usage", request.URL.String())
	require.Equal(t, "Bearer secret", request.Header.Get("Authorization"))
	require.Equal(t, "sub2api-upstream-balance/1.0", request.Header.Get("User-Agent"))
}
