package handler

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// ModelPricingHandler 处理用户侧「模型价格」页面查询。
//
// 对所有登录用户开放（不受 available-channels 开关控制），按用户当前可用分组返回：
//  1. 分组基础信息与倍率（含用户专属倍率、高峰倍率规则）；
//  2. 分组实际可用的模型列表，与网关 /v1/models 的计算逻辑一致
//     （分组自定义列表 → 账号模型映射 → 平台默认列表）；
//  3. 每个模型的基础定价，走与网关计费相同的 ModelPricingResolver
//     （渠道定价 → LiteLLM → 内置 fallback），前端只负责乘倍率展示。
type ModelPricingHandler struct {
	apiKeyService   *service.APIKeyService
	gatewayService  *service.GatewayService
	pricingResolver *service.ModelPricingResolver
	pricingService  *service.PricingService
}

// NewModelPricingHandler 创建用户侧模型价格 handler。
func NewModelPricingHandler(
	apiKeyService *service.APIKeyService,
	gatewayService *service.GatewayService,
	pricingResolver *service.ModelPricingResolver,
	pricingService *service.PricingService,
) *ModelPricingHandler {
	return &ModelPricingHandler{
		apiKeyService:   apiKeyService,
		gatewayService:  gatewayService,
		pricingResolver: pricingResolver,
		pricingService:  pricingService,
	}
}

// modelPricingGroup 用户可见的分组 + 模型价格（白名单字段）。
type modelPricingGroup struct {
	ID                 int64              `json:"id"`
	Name               string             `json:"name"`
	Description        string             `json:"description"`
	Platform           string             `json:"platform"`
	SubscriptionType   string             `json:"subscription_type"`
	RateMultiplier     float64            `json:"rate_multiplier"`
	UserRateMultiplier *float64           `json:"user_rate_multiplier"`
	PeakRateEnabled    bool               `json:"peak_rate_enabled"`
	PeakStart          string             `json:"peak_start"`
	PeakEnd            string             `json:"peak_end"`
	PeakRateMultiplier float64            `json:"peak_rate_multiplier"`
	PeakFactorNow      float64            `json:"peak_factor_now"` // 服务端时区下当前时刻的高峰因子，非高峰为 1
	IsExclusive        bool               `json:"is_exclusive"`
	ModelSource        string             `json:"model_source"` // group / accounts / default
	Models             []modelPricingItem `json:"models"`
}

// modelPricingItem 单个模型的基础定价（倍率前，USD / token）。
type modelPricingItem struct {
	Name            string  `json:"name"`
	BillingMode     string  `json:"billing_mode"`   // token / per_request / image / video
	PricingSource   string  `json:"pricing_source"` // group / channel / litellm / fallback
	Priced          bool    `json:"priced"`
	InputPrice      float64 `json:"input_price"`
	OutputPrice     float64 `json:"output_price"`
	CacheWritePrice float64 `json:"cache_write_price"`
	CacheReadPrice  float64 `json:"cache_read_price"`
	PerRequestPrice float64 `json:"per_request_price"`

	Intervals []modelPricingInterval `json:"intervals,omitempty"`

	LongContextThreshold        int     `json:"long_context_threshold,omitempty"`
	LongContextInputMultiplier  float64 `json:"long_context_input_multiplier,omitempty"`
	LongContextOutputMultiplier float64 `json:"long_context_output_multiplier,omitempty"`

	ContextWindow int  `json:"context_window,omitempty"`
	Reasoning     bool `json:"reasoning,omitempty"`
	Vision        bool `json:"vision,omitempty"`
}

// modelPricingInterval 渠道区间 / 层级定价（USD / token 或 USD / 次）。
type modelPricingInterval struct {
	MinTokens       int      `json:"min_tokens"`
	MaxTokens       *int     `json:"max_tokens"`
	TierLabel       string   `json:"tier_label,omitempty"`
	InputPrice      *float64 `json:"input_price"`
	OutputPrice     *float64 `json:"output_price"`
	CacheWritePrice *float64 `json:"cache_write_price"`
	CacheReadPrice  *float64 `json:"cache_read_price"`
	PerRequestPrice *float64 `json:"per_request_price"`
}

// List 列出当前用户可用分组下的模型价格。
// GET /api/v1/model-pricing
func (h *ModelPricingHandler) List(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	ctx := c.Request.Context()

	groups, err := h.apiKeyService.GetAvailableGroups(ctx, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	userRates, err := h.apiKeyService.GetUserGroupRates(ctx, subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	now := time.Now()
	out := make([]modelPricingGroup, 0, len(groups))
	for i := range groups {
		g := &groups[i]
		if g.Status != "" && g.Status != service.StatusActive {
			continue
		}
		modelIDs, source := h.groupModelIDs(ctx, g)
		item := modelPricingGroup{
			ID:                 g.ID,
			Name:               g.Name,
			Description:        g.Description,
			Platform:           g.Platform,
			SubscriptionType:   g.SubscriptionType,
			RateMultiplier:     g.RateMultiplier,
			PeakRateEnabled:    g.PeakRateEnabled && g.IsSubscriptionType(),
			PeakStart:          g.PeakStart,
			PeakEnd:            g.PeakEnd,
			PeakRateMultiplier: g.PeakRateMultiplier,
			PeakFactorNow:      g.PeakMultiplierAt(now),
			IsExclusive:        g.IsExclusive,
			ModelSource:        source,
			Models:             h.priceModels(ctx, g, modelIDs),
		}
		if rate, ok := userRates[g.ID]; ok {
			r := rate
			item.UserRateMultiplier = &r
		}
		out = append(out, item)
	}

	response.Success(c, out)
}

// groupModelIDs 计算分组对用户暴露的模型列表，与 GatewayHandler.Models 保持一致：
// 分组启用模型白名单时按白名单过滤候选列表；否则取分组内账号的模型映射；都没有则回退平台默认列表。
// 返回模型 ID 列表与来源标识（group / accounts / default）。通配符条目无法定价，直接跳过。
func (h *ModelPricingHandler) groupModelIDs(ctx context.Context, g *service.Group) ([]string, string) {
	groupID := g.ID
	platform := g.Platform

	var available []string
	if h.gatewayService != nil {
		if platform == service.PlatformComposite {
			available = compositeAvailableModelsFrom(ctx, h.gatewayService, &groupID, "", true)
		} else {
			available = h.gatewayService.GetAvailableModels(ctx, &groupID, platform)
		}
	}

	if g.ModelAllowlistEnabled() {
		fallback := defaultModelIDsForPlatform(platform)
		source := available
		if platform != service.PlatformComposite {
			source = modelListingSource(platform, available, fallback)
		} else if len(source) == 0 {
			source = fallback
		}
		return concreteModelIDs(g.ModelAllowlist.FilterForListing(source)), "group"
	}
	if len(available) > 0 {
		return concreteModelIDs(available), "accounts"
	}
	return concreteModelIDs(defaultPricingModelIDs(platform)), "default"
}

// defaultPricingModelIDs 平台默认模型列表。Anthropic 与 /v1/models 的兜底一致只用 Claude 列表；
// 其他平台用 defaultModelIDsForPlatform（CN 供应商没有静态列表时返回空，不误展示 Claude 模型）。
func defaultPricingModelIDs(platform string) []string {
	if platform == service.PlatformAnthropic {
		ids := make([]string, 0, len(claude.DefaultModels))
		for _, m := range claude.DefaultModels {
			ids = append(ids, m.ID)
		}
		return ids
	}
	return defaultModelIDsForPlatform(platform)
}

// concreteModelIDs 去空、去重、去通配符。
func concreteModelIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" || strings.Contains(id, "*") {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// priceModels 用与网关计费相同的解析链给每个模型定价（传入分组以命中分组级定价与长上下文开关）。
func (h *ModelPricingHandler) priceModels(ctx context.Context, g *service.Group, modelIDs []string) []modelPricingItem {
	out := make([]modelPricingItem, 0, len(modelIDs))
	for _, name := range modelIDs {
		item := modelPricingItem{Name: name, BillingMode: string(service.BillingModeToken)}
		if h.pricingResolver != nil {
			gid := g.ID
			fillResolvedPricing(&item, h.pricingResolver.Resolve(ctx, service.PricingInput{Model: name, GroupID: &gid, Group: g}))
		}
		if !g.LongContextPricingEnabled {
			hideLongContextPricing(&item)
		}
		if h.pricingService != nil {
			if lp := h.pricingService.GetModelPricing(name); lp != nil {
				item.ContextWindow = lp.MaxInputTokens
				item.Reasoning = lp.SupportsReasoning
				item.Vision = lp.SupportsVision
			}
		}
		out = append(out, item)
	}
	return out
}

// fillResolvedPricing 把解析结果转成展示字段。
func fillResolvedPricing(item *modelPricingItem, rp *service.ResolvedPricing) {
	if rp == nil {
		return
	}
	item.PricingSource = rp.Source
	if rp.Mode != "" {
		item.BillingMode = string(rp.Mode)
	}

	if rp.Mode == service.BillingModePerRequest || rp.Mode == service.BillingModeImage || rp.Mode == service.BillingModeVideo {
		item.PerRequestPrice = rp.DefaultPerRequestPrice
		item.Intervals = toModelPricingIntervals(rp.RequestTiers)
		// 已解析的按次定价允许零价，与网关实际计费保持一致。
		item.Priced = true
		return
	}

	item.Intervals = toModelPricingIntervals(rp.Intervals)
	if bp := rp.BasePricing; bp != nil {
		item.InputPrice = bp.InputPricePerToken
		item.OutputPrice = bp.OutputPricePerToken
		item.CacheWritePrice = bp.CacheCreationPricePerToken
		item.CacheReadPrice = bp.CacheReadPricePerToken
		if bp.LongContextInputThreshold > 0 && (bp.LongContextInputMultiplier > 1 || bp.LongContextOutputMultiplier > 1) {
			item.LongContextThreshold = bp.LongContextInputThreshold
			item.LongContextInputMultiplier = bp.LongContextInputMultiplier
			item.LongContextOutputMultiplier = bp.LongContextOutputMultiplier
		}
	}
	item.Priced = rp.BasePricing != nil || len(item.Intervals) > 0
}

// hideLongContextPricing 分组关闭长上下文计价时，网关计费只取最低档，展示同步隐藏阶梯与长上下文倍率。
func hideLongContextPricing(item *modelPricingItem) {
	item.LongContextThreshold = 0
	item.LongContextInputMultiplier = 0
	item.LongContextOutputMultiplier = 0
	if item.BillingMode != string(service.BillingModeToken) || len(item.Intervals) <= 1 {
		return
	}
	lowest := item.Intervals[0]
	for _, iv := range item.Intervals[1:] {
		if iv.MinTokens < lowest.MinTokens {
			lowest = iv
		}
	}
	item.Intervals = []modelPricingInterval{lowest}
}

func toModelPricingIntervals(src []service.PricingInterval) []modelPricingInterval {
	if len(src) == 0 {
		return nil
	}
	out := make([]modelPricingInterval, 0, len(src))
	for _, iv := range src {
		out = append(out, modelPricingInterval{
			MinTokens:       iv.MinTokens,
			MaxTokens:       iv.MaxTokens,
			TierLabel:       iv.TierLabel,
			InputPrice:      iv.InputPrice,
			OutputPrice:     iv.OutputPrice,
			CacheWritePrice: iv.CacheWritePrice,
			CacheReadPrice:  iv.CacheReadPrice,
			PerRequestPrice: iv.PerRequestPrice,
		})
	}
	return out
}
