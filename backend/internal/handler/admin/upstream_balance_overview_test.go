package admin

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUpstreamBalanceRepresentativesGroupsAccountsByCanonicalBaseURL(t *testing.T) {
	accounts := []service.Account{
		{ID: 3, Type: service.AccountTypeAPIKey, Status: service.StatusActive, Credentials: map[string]any{"api_key": "first", "base_url": "https://same.example/"}},
		{ID: 8, Type: service.AccountTypeAPIKey, Status: service.StatusActive, Credentials: map[string]any{"api_key": "second", "base_url": "https://same.example"}},
		{ID: 11, Type: service.AccountTypeAPIKey, Status: service.StatusActive, Credentials: map[string]any{"api_key": "other", "base_url": "https://other.example/v1/"}},
		{ID: 15, Type: service.AccountTypeAPIKey, Status: "inactive", Credentials: map[string]any{"api_key": "inactive", "base_url": "https://inactive.example"}},
	}

	got := upstreamBalanceRepresentatives(accounts)
	require.Equal(t, []upstreamBalanceRepresentative{
		{baseURL: "https://other.example/v1", accountID: 11},
		{baseURL: "https://same.example", accountID: 3},
	}, got)
}
