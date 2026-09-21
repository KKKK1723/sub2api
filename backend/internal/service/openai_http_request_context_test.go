//go:build unit

package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenAIHTTPContextCancellationLifecycle(t *testing.T) {
	t.Run("already_canceled", func(t *testing.T) {
		parent, cancel := context.WithCancel(context.Background())
		cancel()
		ctx, cleanup := newOpenAIHTTPUpstreamContext(parent)
		defer cleanup()
		require.ErrorIs(t, ctx.Err(), context.Canceled)
		httptrace.ContextClientTrace(ctx).GotConn(httptrace.GotConnInfo{})
		require.ErrorIs(t, ctx.Err(), context.Canceled)
	})
	t.Run("deadline_while_waiting", func(t *testing.T) {
		parent, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()
		ctx, cleanup := newOpenAIHTTPUpstreamContext(parent)
		defer cleanup()
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
			t.Fatal("取连接期间未继承客户端期限的取消")
		}
	})
	t.Run("connected_retains_values_and_traces", func(t *testing.T) {
		type key struct{}
		var existingTraceCalls atomic.Int32
		parent, cancel := context.WithCancel(context.WithValue(context.Background(), key{}, "value"))
		defer cancel()
		parent = httptrace.WithClientTrace(parent, &httptrace.ClientTrace{
			GotConn: func(httptrace.GotConnInfo) { existingTraceCalls.Add(1) },
		})
		ctx, cleanup := newOpenAIHTTPUpstreamContext(parent)
		defer cleanup()
		require.Equal(t, "value", ctx.Value(key{}))
		trace := httptrace.ContextClientTrace(ctx)
		trace.GotConn(httptrace.GotConnInfo{})
		cancel()
		// 重定向或透明重试再次取连接时，不能恢复已解除的客户端取消联动。
		trace.GotConn(httptrace.GotConnInfo{})
		require.NoError(t, ctx.Err())
		require.EqualValues(t, 2, existingTraceCalls.Load())
		cleanup()
		require.ErrorIs(t, ctx.Err(), context.Canceled)
	})
	t.Run("nil_parent", func(t *testing.T) {
		ctx, cleanup := newOpenAIHTTPUpstreamContext(nil)
		require.NoError(t, ctx.Err())
		cleanup()
		require.ErrorIs(t, ctx.Err(), context.Canceled)
	})
}

type lifecycleHTTPUpstream struct {
	HTTPUpstream
	client   *http.Client
	contexts chan context.Context
}

func (u *lifecycleHTTPUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	if u.contexts != nil {
		u.contexts <- req.Context()
	}
	return u.client.Do(req)
}

// 使用真实 HTTP/1.1 连接池走完整 Forward：取消排队请求，同时保留已发出请求的用量。
func TestOpenAIForwardConnectionWaitCancellationAndUsageDrain(t *testing.T) {
	for _, mode := range []string{"native", "native_guard", "passthrough"} {
		t.Run(mode, func(t *testing.T) {
			firstArrived := make(chan struct{})
			unblockFirst := make(chan struct{})
			var unblockOnce sync.Once
			unblock := func() { unblockOnce.Do(func() { close(unblockFirst) }) }
			var requestCount atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.Copy(io.Discard, r.Body)
				if requestCount.Add(1) == 1 {
					close(firstArrived)
					<-unblockFirst
				}
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = io.WriteString(w, "data: {\"type\":\"response.output_text.delta\",\"delta\":\"hello\"}\n\n"+
					"data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_lifecycle\",\"usage\":{\"input_tokens\":10,\"output_tokens\":5}}}\n\n")
			}))
			defer server.Close()
			defer unblock()
			transport := &http.Transport{MaxConnsPerHost: 1}
			defer transport.CloseIdleConnections()
			svc := newOpenAIImageGenerationControlTestService(nil)
			contexts := make(chan context.Context, 3)
			svc.httpUpstream = &lifecycleHTTPUpstream{
				client: &http.Client{Transport: transport, Timeout: 5 * time.Second}, contexts: contexts,
			}
			svc.cfg.Security.URLAllowlist.AllowInsecureHTTP = true
			if mode == "native_guard" {
				svc.cfg.Gateway.OpenAIFirstOutputTimeoutSeconds = 3
			}
			account := newOpenAIImageGenerationControlTestAccount()
			account.Credentials["base_url"] = server.URL
			account.Extra = map[string]any{"openai_responses_supported": true, "openai_passthrough": mode == "passthrough"}
			body := []byte(`{"model":"gpt-5.6-sol","stream":true,"input":"hello"}`)
			type outcome struct {
				result *OpenAIForwardResult
				err    error
			}
			forward := func(parent context.Context) <-chan outcome {
				c, _ := newOpenAIImageGenerationControlTestContext(true, "codex_cli_rs/0.144.1")
				c.Request = c.Request.WithContext(parent)
				done := make(chan outcome, 1)
				go func() {
					result, err := svc.Forward(parent, c, account, body)
					done <- outcome{result, err}
				}()
				return done
			}
			await := func(done <-chan outcome) outcome {
				t.Helper()
				select {
				case result := <-done:
					return result
				case <-time.After(2 * time.Second):
					t.Fatal("请求未在预期时间结束")
					return outcome{}
				}
			}
			firstParent, cancelFirst := context.WithCancel(context.Background())
			defer cancelFirst()
			firstDone := forward(firstParent)
			select {
			case <-firstArrived:
			case <-time.After(2 * time.Second):
				t.Fatal("首个请求未到达上游")
			}
			cancelFirst()
			secondParent, cancelSecond := context.WithCancel(context.Background())
			defer cancelSecond()
			gettingConn := make(chan struct{}, 1)
			var secondGotConn atomic.Bool
			secondParent = httptrace.WithClientTrace(secondParent, &httptrace.ClientTrace{
				GetConn: func(string) { gettingConn <- struct{}{} },
				GotConn: func(httptrace.GotConnInfo) { secondGotConn.Store(true) },
			})
			secondDone := forward(secondParent)
			select {
			case <-gettingConn:
			case <-time.After(2 * time.Second):
				t.Fatal("第二个请求没有进入连接池")
			}
			require.False(t, secondGotConn.Load())
			cancelSecond()
			second := await(secondDone)
			require.ErrorIs(t, second.err, context.Canceled)
			var failover *UpstreamFailoverError
			require.NotErrorAs(t, second.err, &failover)
			require.EqualValues(t, 1, requestCount.Load())
			select {
			case <-firstDone:
				t.Fatal("已发送的请求被客户端取消，无法继续收集用量")
			default:
			}
			unblock()
			first := await(firstDone)
			require.NoError(t, first.err)
			require.NotNil(t, first.result)
			require.Equal(t, 10, first.result.Usage.InputTokens)
			require.Equal(t, 5, first.result.Usage.OutputTokens)
			third := await(forward(context.Background()))
			require.NoError(t, third.err)
			// 第三个请求可以正常发送；已取消的第二个请求始终没有到达上游。
			require.EqualValues(t, 2, requestCount.Load())
			require.False(t, secondGotConn.Load())
			for range 3 {
				require.ErrorIs(t, (<-contexts).Err(), context.Canceled, "转发结束后必须清理上游上下文")
			}
		})
	}
}
