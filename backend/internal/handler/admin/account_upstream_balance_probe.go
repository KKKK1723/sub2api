package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type upstreamBalanceProbeBatchRequest struct {
	AccountIDs []int64 `json:"account_ids" binding:"required"`
}

func (h *AccountHandler) ProbeUpstreamBalance(c *gin.Context) {
	if h.upstreamBillingProbe == nil {
		response.ErrorFrom(c, service.ErrUpstreamBalanceProbeUnavailable)
		return
	}
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "Invalid account ID")
		return
	}
	snapshot, err := h.upstreamBillingProbe.ProbeUpstreamBalance(c.Request.Context(), accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, service.UpstreamBalanceProbeResult{AccountID: accountID, Snapshot: snapshot})
}

func (h *AccountHandler) ProbeUpstreamBalanceBatch(c *gin.Context) {
	if h.upstreamBillingProbe == nil {
		response.ErrorFrom(c, service.ErrUpstreamBalanceProbeUnavailable)
		return
	}
	var req upstreamBalanceProbeBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if len(req.AccountIDs) == 0 || len(req.AccountIDs) > service.UpstreamBillingProbeMaxBatchSize {
		response.BadRequest(c, "account_ids must contain between 1 and 20 items")
		return
	}
	seen := make(map[int64]struct{}, len(req.AccountIDs))
	accountIDs := make([]int64, 0, len(req.AccountIDs))
	for _, accountID := range req.AccountIDs {
		if accountID <= 0 {
			response.BadRequest(c, "account_ids must contain positive IDs")
			return
		}
		if _, exists := seen[accountID]; exists {
			continue
		}
		seen[accountID] = struct{}{}
		accountIDs = append(accountIDs, accountID)
	}
	response.Success(c, gin.H{"results": h.upstreamBillingProbe.ProbeUpstreamBalances(c.Request.Context(), accountIDs)})
}
