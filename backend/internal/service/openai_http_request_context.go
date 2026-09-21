package service

import (
	"context"
	"net/http/httptrace"
	"sync"
)

// newOpenAIHTTPUpstreamContext 在获取首个连接前保留客户端取消能力。
// 取得连接后可能立即发送请求，此时脱离客户端取消以继续收集用量；
// 调用方必须在请求失败或响应体关闭时调用 cleanup，回收请求上下文。
func newOpenAIHTTPUpstreamContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(context.WithoutCancel(parent))
	var mu sync.Mutex
	connected := false
	cancelWaiting := func() {
		mu.Lock()
		defer mu.Unlock()
		if !connected {
			cancel()
		}
	}
	stop := context.AfterFunc(parent, cancelWaiting)
	// AfterFunc 异步执行，已取消的请求必须在返回前同步失效。
	if parent.Err() != nil {
		cancelWaiting()
	}
	ctx = httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		GotConn: func(httptrace.GotConnInfo) {
			// 与取消回调串行化，确保取消已获胜时先取消上下文再返回 Transport。
			mu.Lock()
			if !connected {
				if parent.Err() != nil {
					cancel()
				} else {
					connected = true
				}
			}
			mu.Unlock()
			stop()
		},
	})
	return ctx, func() {
		stop()
		cancel()
	}
}
