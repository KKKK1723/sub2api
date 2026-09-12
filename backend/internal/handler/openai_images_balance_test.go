//go:build unit

package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// 上游图片账号余额不足与其他上游余额问题一致：客户端统一提示分组暂不可用，不暴露上游余额原因。
func TestOpenAIImagesBalanceExhaustionHidesUpstreamBalanceCause(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, streamStarted := range []bool{false, true} {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/edits", nil)
		(&OpenAIGatewayHandler{}).handleFailoverExhausted(c, &service.UpstreamFailoverError{
			StatusCode:       http.StatusBadGateway,
			Reason:           service.OpenAIImagesInsufficientBalanceReason,
			ClientStatusCode: http.StatusPaymentRequired,
			ClientMessage:    service.OpenAIImagesInsufficientBalanceMessage,
			ResponseHeaders:  http.Header{"Retry-After": []string{"23"}, "X-Request-Id": []string{"req-images-balance"}},
		}, streamStarted)
		require.Equal(t, "23", recorder.Header().Get("Retry-After"))
		require.NotContains(t, recorder.Body.String(), service.OpenAIImagesInsufficientBalanceCode)
		require.NotContains(t, recorder.Body.String(), service.OpenAIImagesInsufficientBalanceMessage)

		if streamStarted {
			require.Contains(t, recorder.Body.String(), service.UpstreamGroupUnavailableClientMessage)
		} else {
			require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
			require.Equal(t, service.UpstreamGroupUnavailableClientMessage, gjson.GetBytes(recorder.Body.Bytes(), "error.message").String())
		}
	}
}
