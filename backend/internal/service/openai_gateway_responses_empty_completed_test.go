//go:build unit

package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// 空完成事件应触发故障转移，不能给客户端返回成功的空响应。
func TestOpenAIResponsesEmptyCompletedFailsOver(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, mode := range []string{"native", "native_guard", "passthrough"} {
		t.Run(mode, func(t *testing.T) {
			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
				Body: io.NopCloser(strings.NewReader(
					"data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_empty\",\"object\":\"response\",\"status\":\"in_progress\"}}\n\n" +
						"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_empty\",\"object\":\"response\",\"status\":\"completed\"}}\n\n",
				)),
			}}
			svc := newOpenAIImageGenerationControlTestService(upstream)
			c, recorder := newOpenAIImageGenerationControlTestContext(true, "codex_cli_rs/0.144.1")
			account := newOpenAIImageGenerationControlTestAccount()
			account.Extra = map[string]any{"openai_passthrough": mode == "passthrough", "openai_responses_supported": true}
			if mode == "native_guard" {
				svc.cfg.Gateway.OpenAIFirstOutputTimeoutSeconds = 1
			}

			body := []byte(`{
				"model":"gpt-5.6-sol",
				"stream":true,
				"input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"continue"}]}]
			}`)

			_, err := svc.Forward(context.Background(), c, account, body)
			require.Error(t, err)
			var failoverErr *UpstreamFailoverError
			require.True(t, errors.As(err, &failoverErr), "empty completed must produce UpstreamFailoverError, got: %v", err)
			require.Equal(t, http.StatusBadGateway, failoverErr.StatusCode)
			require.Empty(t, recorder.Body.String(), "no empty success stream may reach the client")
		})
	}
}

// 已经产生有效输出的流应正常完成。
func TestOpenAIResponsesEmptyCompletedWithOutputSucceeds(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, mode := range []string{"native", "native_guard", "passthrough"} {
		t.Run(mode, func(t *testing.T) {
			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
				Body: io.NopCloser(strings.NewReader(
					"data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_ok\",\"object\":\"response\",\"status\":\"in_progress\"}}\n\n" +
						"data: {\"type\":\"response.output_text.delta\",\"delta\":\"hello\"}\n\n" +
						"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_ok\",\"object\":\"response\",\"status\":\"completed\",\"usage\":{\"input_tokens\":10,\"output_tokens\":5,\"total_tokens\":15}}}\n\n",
				)),
			}}
			svc := newOpenAIImageGenerationControlTestService(upstream)
			c, recorder := newOpenAIImageGenerationControlTestContext(true, "codex_cli_rs/0.144.1")
			account := newOpenAIImageGenerationControlTestAccount()
			account.Extra = map[string]any{"openai_passthrough": mode == "passthrough", "openai_responses_supported": true}
			if mode == "native_guard" {
				svc.cfg.Gateway.OpenAIFirstOutputTimeoutSeconds = 1
			}

			body := []byte(`{
				"model":"gpt-5.6-sol",
				"stream":true,
				"input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"continue"}]}]
			}`)

			result, err := svc.Forward(context.Background(), c, account, body)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Contains(t, recorder.Body.String(), "hello")
			require.NotNil(t, result.Usage)
			require.Equal(t, 10, result.Usage.InputTokens)
			require.Equal(t, 5, result.Usage.OutputTokens)
		})
	}
}

// 即使没有文字，只要上游报告了用量就不能作为空响应重试。
func TestOpenAIResponsesEmptyCompletedWithUsageSucceeds(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, mode := range []string{"native", "native_guard", "passthrough"} {
		t.Run(mode, func(t *testing.T) {
			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
				Body: io.NopCloser(strings.NewReader(
					"data: {\"type\":\"response.created\",\"response\":{\"id\":\"resp_usage\",\"object\":\"response\",\"status\":\"in_progress\"}}\n\n" +
						"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_usage\",\"object\":\"response\",\"status\":\"completed\",\"usage\":{\"input_tokens\":3,\"output_tokens\":0,\"total_tokens\":3}}}\n\n",
				)),
			}}
			svc := newOpenAIImageGenerationControlTestService(upstream)
			c, _ := newOpenAIImageGenerationControlTestContext(true, "codex_cli_rs/0.144.1")
			account := newOpenAIImageGenerationControlTestAccount()
			account.Extra = map[string]any{"openai_passthrough": mode == "passthrough", "openai_responses_supported": true}
			if mode == "native_guard" {
				svc.cfg.Gateway.OpenAIFirstOutputTimeoutSeconds = 1
			}

			body := []byte(`{
				"model":"gpt-5.6-sol",
				"stream":true,
				"input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"continue"}]}]
			}`)

			result, err := svc.Forward(context.Background(), c, account, body)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotNil(t, result.Usage)
			require.Equal(t, 3, result.Usage.InputTokens)
		})
	}
}

func TestOpenAIResponsesCompletedEventIsEmpty(t *testing.T) {
	cases := []struct {
		name  string
		data  string
		usage *OpenAIUsage
		want  bool
	}{
		{
			name: "bare completed",
			data: `{"type":"response.completed"}`,
			want: true,
		},
		{
			name: "completed with empty output array",
			data: `{"type":"response.completed","response":{"id":"r1","status":"completed","output":[]}}`,
			want: true,
		},
		{
			name: "completed with usage",
			data: `{"type":"response.completed","response":{"id":"r1","status":"completed","usage":{"input_tokens":1,"output_tokens":1}}}`,
			want: false,
		},
		{
			name: "completed with error",
			data: `{"type":"response.completed","response":{"id":"r1","status":"completed","error":{"code":"x"}}}`,
			want: false,
		},
		{
			name: "completed with output item",
			data: `{"type":"response.completed","response":{"id":"r1","status":"completed","output":[{"type":"message","id":"msg_1"}]}}`,
			want: false,
		},
		{
			name: "accumulated usage",
			data: `{"type":"response.completed"}`,
			usage: &OpenAIUsage{
				InputTokens:  7,
				OutputTokens: 2,
			},
			want: false,
		},
		{
			name: "invalid json",
			data: `{"type":`,
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, openAIResponsesCompletedEventIsEmpty([]byte(tc.data), tc.usage))
		})
	}
}
