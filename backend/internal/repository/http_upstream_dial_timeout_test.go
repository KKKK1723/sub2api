package repository

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// 回归：上游 Transport 必须显式配置建连超时。
//
// http.Transport.DialContext 为 nil 时 Go 使用零值 net.Dialer（Timeout=0），
// DNS 解析与 TCP 握手没有任何上限，只能等内核重传耗尽（Linux 约 130 秒）。
// ResponseHeaderTimeout 只覆盖连接建立之后的阶段，管不到建连。
// 上游域名被解析到不可达 IP 时，串行的多账号故障转移会把一次请求拖到数分钟。
func TestBuildUpstreamTransportSetsDialTimeout(t *testing.T) {
	settings := defaultPoolSettings(nil)

	transport, err := buildUpstreamTransport(settings, nil, upstreamProtocolModeDefault)
	require.NoError(t, err)
	require.NotNil(t, transport.DialContext, "DialContext 缺失会退化为无超时的零值 dialer")
	require.Equal(t, defaultUpstreamTLSHandshakeTimeout, transport.TLSHandshakeTimeout)
}

func TestNewUpstreamDialerHasBoundedTimeout(t *testing.T) {
	dialer := newUpstreamDialer()

	require.Greater(t, dialer.Timeout, time.Duration(0), "建连超时必须有上限")
	require.Equal(t, defaultUpstreamDialTimeout, dialer.Timeout)
	require.Equal(t, defaultUpstreamDialKeepAlive, dialer.KeepAlive)
}

// 建连超时对 HTTP 代理同样生效：Transport.Proxy 走的仍是 DialContext，
// 代理地址不可达时必须快速失败而不是挂满内核超时。
func TestBuildUpstreamTransportKeepsDialTimeoutWithHTTPProxy(t *testing.T) {
	proxyURL, err := url.Parse("http://127.0.0.1:1080")
	require.NoError(t, err)

	transport, err := buildUpstreamTransport(defaultPoolSettings(nil), proxyURL, upstreamProtocolModeDefault)
	require.NoError(t, err)
	require.NotNil(t, transport.Proxy)
	require.NotNil(t, transport.DialContext)
}

// SOCKS5 分支会覆盖 Transport.DialContext，覆盖后仍必须是有超时的拨号器。
func TestBuildUpstreamTransportKeepsDialContextWithSOCKS5Proxy(t *testing.T) {
	proxyURL, err := url.Parse("socks5h://127.0.0.1:1080")
	require.NoError(t, err)

	transport, err := buildUpstreamTransport(defaultPoolSettings(nil), proxyURL, upstreamProtocolModeDefault)
	require.NoError(t, err)
	require.NotNil(t, transport.DialContext)
}

// 已取消的请求必须立即中止拨号，不依赖外网或内核重传超时。
func TestUpstreamDialerRespectsContextCancellation(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()
	require.NoError(t, listener.Close())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	conn, err := newUpstreamDialer().DialContext(ctx, "tcp", addr)
	if conn != nil {
		_ = conn.Close()
	}
	require.ErrorIs(t, err, context.Canceled, "已取消的 context 必须立即中止拨号")
}

// TCP 已连接但对端不推进 TLS 握手时，应由握手超时主动结束请求。
func TestUpstreamTransportStopsStalledTLSHandshake(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	release := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		<-release
	}()
	t.Cleanup(func() {
		close(release)
		_ = listener.Close()
		<-done
	})

	transport, err := buildUpstreamTransport(defaultPoolSettings(nil), nil, upstreamProtocolModeDefault)
	require.NoError(t, err)
	// 缩短测试等待；生产默认值由上面的配置回归测试单独核对。
	transport.TLSHandshakeTimeout = 80 * time.Millisecond
	t.Cleanup(transport.CloseIdleConnections)
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}
	resp, err := client.Get("https://" + listener.Addr().String())
	if resp != nil {
		_ = resp.Body.Close()
	}
	require.ErrorContains(t, err, "TLS handshake timeout", "不应依赖整个请求的兜底超时")
}
