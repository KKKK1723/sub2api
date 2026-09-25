//go:build unit

package service

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMonitorPlatformRequests(t *testing.T) {
	swapMonitorHTTPClient(t)
	for _, tc := range []struct{ provider, base, mode, path string }{
		{"kimi", "/v1", "chat_completions", "/v1/chat/completions"},
		{"zhipu", "", "chat_completions", "/api/paas/v4/chat/completions"},
		{"zhipu", "/api/coding/paas/v4", "chat_completions", "/api/coding/paas/v4/chat/completions"},
		{"deepseek", "", "responses", "/responses"},
		{"deepseek", "/v1", "chat_completions", "/v1/chat/completions"},
		{"minimax", "/v1", "responses", "/v1/responses"},
		{"opencode_go", "/zen/go/v1", "chat_completions", "/zen/go/v1/chat/completions"},
		{"opencode_go", "/zen/v1", "responses", "/zen/v1/responses"},
	} {
		t.Run(tc.provider+tc.base+tc.mode, func(t *testing.T) {
			handler := &openAICaptureHandler{rawResponse: `{"choices":[{"message":{"content":"ok"}}]}`}
			server := httptest.NewServer(handler)
			defer server.Close()
			_, _, status, err := callProvider(context.Background(), tc.provider, server.URL+tc.base, "test-key", "model", "ping", &CheckOptions{APIMode: tc.mode})
			require.NoError(t, err)
			require.Equal(t, 200, status)
			require.Equal(t, tc.path, handler.lastPath)
			require.Equal(t, "Bearer test-key", handler.lastHeaders.Get("Authorization"))
			if tc.provider == "opencode_go" {
				require.NotEmpty(t, handler.lastHeaders.Get("X-OpenCode-Session"))
			}
		})
	}
}

func TestMonitorCNBodyValidation(t *testing.T) {
	for _, provider := range []string{"kimi", "zhipu", "deepseek", "minimax", "opencode_go"} {
		require.Error(t, validateReplaceRequestBody(provider, "chat_completions", map[string]any{"input": "wrong protocol"}))
		require.Error(t, validateReplaceRequestBody(provider, "responses", map[string]any{"messages": "wrong protocol"}))
		require.Equal(t, bodyMergeDenyKey("openai", "responses"), bodyMergeDenyKey(provider, "responses"))
	}
}
