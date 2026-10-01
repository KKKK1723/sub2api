//go:build unit

package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestModelPricing_Unauthenticated401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ModelPricingHandler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/model-pricing", nil)

	h.List(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestConcreteModelIDs_DedupSortAndDropWildcards(t *testing.T) {
	got := concreteModelIDs([]string{" gpt-5 ", "claude-*", "", "gpt-5", "claude-opus-4-6"})
	require.Equal(t, []string{"claude-opus-4-6", "gpt-5"}, got)
}

func TestFillResolvedPricing_Token(t *testing.T) {
	item := modelPricingItem{}
	fillResolvedPricing(&item, &service.ResolvedPricing{
		Mode:   service.BillingModeToken,
		Source: service.PricingSourceLiteLLM,
		BasePricing: &service.ModelPricing{
			InputPricePerToken:          3e-6,
			OutputPricePerToken:         15e-6,
			CacheCreationPricePerToken:  3.75e-6,
			CacheReadPricePerToken:      0.3e-6,
			LongContextInputThreshold:   200000,
			LongContextInputMultiplier:  2,
			LongContextOutputMultiplier: 1.5,
		},
	})
	require.True(t, item.Priced)
	require.Equal(t, "token", item.BillingMode)
	require.Equal(t, "litellm", item.PricingSource)
	require.InDelta(t, 3e-6, item.InputPrice, 1e-12)
	require.InDelta(t, 15e-6, item.OutputPrice, 1e-12)
	require.Equal(t, 200000, item.LongContextThreshold)
	require.InDelta(t, 2.0, item.LongContextInputMultiplier, 1e-9)
}

func TestFillResolvedPricing_PerRequest(t *testing.T) {
	price := 0.04
	item := modelPricingItem{}
	fillResolvedPricing(&item, &service.ResolvedPricing{
		Mode:                   service.BillingModePerRequest,
		Source:                 service.PricingSourceChannel,
		DefaultPerRequestPrice: 0.02,
		RequestTiers:           []service.PricingInterval{{TierLabel: "4K", PerRequestPrice: &price}},
	})
	require.True(t, item.Priced)
	require.Equal(t, "per_request", item.BillingMode)
	require.InDelta(t, 0.02, item.PerRequestPrice, 1e-12)
	require.Len(t, item.Intervals, 1)
	require.Equal(t, "4K", item.Intervals[0].TierLabel)
}

func TestFillResolvedPricing_UnpricedWhenNoBase(t *testing.T) {
	item := modelPricingItem{BillingMode: "token"}
	fillResolvedPricing(&item, &service.ResolvedPricing{Mode: service.BillingModeToken})
	require.False(t, item.Priced)
}

func TestFillResolvedPricing_ZeroAndCacheOnlyPricesRemainAvailable(t *testing.T) {
	for _, pricing := range []*service.ModelPricing{
		{},
		{CacheReadPricePerToken: 0.3e-6},
	} {
		item := modelPricingItem{}
		fillResolvedPricing(&item, &service.ResolvedPricing{Mode: service.BillingModeToken, BasePricing: pricing})
		require.True(t, item.Priced)
	}
	for _, mode := range []service.BillingMode{service.BillingModePerRequest, service.BillingModeImage} {
		item := modelPricingItem{}
		fillResolvedPricing(&item, &service.ResolvedPricing{Mode: mode})
		require.True(t, item.Priced)
		require.Zero(t, item.PerRequestPrice)
	}
}
