package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type responseAffinityContextKey struct{}

type responseBindContextProbeCache struct {
	stubGatewayCache
	writeContext context.Context
	writeError   error
	writeValue   any
	deadline     time.Time
	writes       int
}

func (c *responseBindContextProbeCache) SetSessionAccountID(ctx context.Context, groupID int64, sessionHash string, accountID int64, ttl time.Duration) error {
	c.writes++
	c.writeContext = ctx
	c.writeError = ctx.Err()
	c.writeValue = ctx.Value(responseAffinityContextKey{})
	c.deadline, _ = ctx.Deadline()
	if c.writeError != nil {
		return c.writeError
	}
	return c.stubGatewayCache.SetSessionAccountID(ctx, groupID, sessionHash, accountID, ttl)
}

func TestOpenAIBindHTTPResponseAccountPreservesAffinityAfterDisconnect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, state := range []string{"active", "canceled", "expired", "nil"} {
		t.Run(state, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			groupID := int64(4201)
			c.Set("api_key", &APIKey{ID: 501, GroupID: &groupID})
			cache := &responseBindContextProbeCache{}
			svc := &OpenAIGatewayService{cache: cache}
			account := &Account{ID: 37001, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}

			var requestCtx context.Context
			if state != "nil" {
				requestCtx = context.WithValue(context.Background(), responseAffinityContextKey{}, "request-value")
				if state == "canceled" {
					var cancel context.CancelFunc
					requestCtx, cancel = context.WithCancel(requestCtx)
					cancel()
				} else if state == "expired" {
					var cancel context.CancelFunc
					requestCtx, cancel = context.WithDeadline(requestCtx, time.Now().Add(-time.Second))
					defer cancel()
				}
			}
			startedAt := time.Now()
			svc.bindHTTPResponseAccount(requestCtx, c, account, "resp_http_canceled_001")

			require.Equal(t, 1, cache.writes)
			require.NoError(t, cache.writeError, "客户端取消不能中止响应关联写入")
			require.True(t, cache.deadline.After(startedAt))
			require.LessOrEqual(t, cache.deadline.Sub(startedAt), openAIWSStateStoreRedisTimeout+100*time.Millisecond)
			if state != "nil" {
				require.Equal(t, "request-value", cache.writeValue, "保留请求上下文的信息")
			}
			require.Equal(t, account.ID, cache.sessionBindings[openAIWSResponseAccountCacheKey("resp_http_canceled_001")])
			require.ErrorIs(t, cache.writeContext.Err(), context.Canceled, "写入完成后应释放独立上下文")
		})
	}
}
