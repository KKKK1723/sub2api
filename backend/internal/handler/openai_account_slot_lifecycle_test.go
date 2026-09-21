package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type canceledOpenAIConnectionUpstream struct {
	service.HTTPUpstream
	cancel context.CancelFunc
	calls  int
}

func (u *canceledOpenAIConnectionUpstream) Do(*http.Request, string, int64, int) (*http.Response, error) {
	u.calls++
	u.cancel()
	return nil, context.Canceled
}

func TestOpenAIResponsesCanceledConnectionDoesNotBecomeUpstreamFailure(t *testing.T) {
	for _, passthrough := range []bool{false, true} {
		name := "native"
		if passthrough {
			name = "passthrough"
		}
		t.Run(name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			cfg := &config.Config{RunMode: config.RunModeSimple}
			cfg.Default.RateMultiplier = 1
			accountRepo := &openAIWSFailoverHandlerAccountRepoStub{accounts: []service.Account{{
				ID: 1, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey,
				Status: service.StatusActive, Schedulable: true,
				Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://example.test"},
				Extra:       map[string]any{"openai_passthrough": passthrough, "openai_responses_supported": true},
			}}}
			parent, cancel := context.WithCancel(context.Background())
			defer cancel()
			upstream := &canceledOpenAIConnectionUpstream{cancel: cancel}
			billingCache := service.NewBillingCacheService(nil, nil, nil, nil, nil, nil, cfg, nil)
			defer billingCache.Stop()
			gateway := service.NewOpenAIGatewayService(
				accountRepo, nil, nil, nil, nil, nil, nil, cfg, nil, nil,
				service.NewBillingService(cfg, nil), nil, billingCache, upstream,
				&service.DeferredService{}, nil, nil, nil, nil, nil, nil, nil,
			)
			h := NewOpenAIGatewayHandler(gateway, service.NewConcurrencyService(nil), billingCache,
				service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, cfg), nil, nil, nil, nil, cfg)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses",
				strings.NewReader(`{"model":"gpt-5.2","input":"hello","stream":true}`)).WithContext(parent)
			groupID := int64(1)
			c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{
				ID: 1, GroupID: &groupID,
				User:  &service.User{ID: 1, Status: service.StatusActive},
				Group: &service.Group{ID: groupID, Platform: service.PlatformOpenAI, Status: service.StatusActive},
			})
			c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1})
			h.Responses(c)
			require.Equal(t, 1, upstream.calls)
			require.Equal(t, statusClientClosedRequest, c.Writer.Status())
			require.Empty(t, rec.Body.String())
			require.Zero(t, gateway.SnapshotOpenAIAccountSchedulerMetrics().RuntimeStatsAccountCount)
		})
	}
}

func TestOpenAIAccountSlotHeldUntilForwardReturns(t *testing.T) {
	for _, mode := range []string{"selected", "fast", "wait"} {
		t.Run(mode, func(t *testing.T) {
			cache := &helperConcurrencyCacheStub{accountSeq: []bool{true}}
			if mode == "wait" {
				cache.accountSeq = []bool{false, true}
			}
			h := &OpenAIGatewayHandler{
				gatewayService: &service.OpenAIGatewayService{},
				concurrencyHelper: NewConcurrencyHelper(
					service.NewConcurrencyService(cache), SSEPingFormatNone, 0,
				),
			}
			parent, cancel := context.WithCancel(context.Background())
			defer cancel()
			c, _ := newHelperTestContext(http.MethodPost, "/v1/responses")
			c.Request = c.Request.WithContext(parent)
			selection := &service.AccountSelectionResult{
				Account: &service.Account{ID: 1},
				WaitPlan: &service.AccountWaitPlan{
					MaxConcurrency: 1, MaxWaiting: 1, Timeout: time.Second,
				},
			}
			if mode == "selected" {
				selection.Acquired = true
				selection.ReleaseFunc = func() { _ = cache.ReleaseAccountSlot(context.Background(), 1, "test") }
			}
			streamStarted := false
			release, acquired := h.acquireResponsesAccountSlot(c, nil, "", selection, false, &streamStarted, zap.NewNop())
			require.True(t, acquired)
			require.NotNil(t, release)
			defer release()
			userReleased := make(chan struct{})
			releaseUser := wrapReleaseOnDone(parent, func() { close(userReleased) })
			defer releaseUser()
			cancel()
			select {
			case <-userReleased:
			case <-time.After(time.Second):
				t.Fatal("客户端取消后用户槽位没有释放")
			}
			// 等待取消回调的调度窗口，账号槽位必须仍被本次上游请求占用。
			require.Never(t, func() bool {
				cache.mu.Lock()
				defer cache.mu.Unlock()
				return cache.accountReleaseCalls != 0
			}, 50*time.Millisecond, time.Millisecond)
			var wg sync.WaitGroup
			for range 8 {
				wg.Go(release)
			}
			wg.Wait()
			cache.mu.Lock()
			count := cache.accountReleaseCalls
			cache.mu.Unlock()
			require.Equal(t, 1, count)
		})
	}
}
