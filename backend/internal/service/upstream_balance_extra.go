package service

import (
	"encoding/json"
	"maps"
)

// NormalizeUpstreamBalanceAccountExtra validates account-managed balance
// configuration and always drops caller-supplied runtime snapshots.
func NormalizeUpstreamBalanceAccountExtra(platform, accountType string, extra map[string]any) (map[string]any, error) {
	normalized := maps.Clone(extra)
	if normalized == nil {
		normalized = make(map[string]any)
	}
	delete(normalized, UpstreamBalanceProbeExtraKey)

	rawEnabled, hasEnabled := normalized[UpstreamBalanceProbeEnabledExtraKey]
	enabled := false
	if hasEnabled {
		var ok bool
		enabled, ok = rawEnabled.(bool)
		if !ok {
			return nil, invalidUpstreamBalanceQuery("upstream_balance_probe_enabled must be a boolean")
		}
	}

	rawQuery, hasQuery := normalized[UpstreamBalanceQueryExtraKey]
	if accountType != AccountTypeAPIKey {
		if enabled || (hasQuery && rawQuery != nil) {
			return nil, ErrUpstreamBalanceProbeAccountInvalid
		}
		delete(normalized, UpstreamBalanceProbeEnabledExtraKey)
		delete(normalized, UpstreamBalanceQueryExtraKey)
		return normalized, nil
	}

	if rawQuery == nil {
		hasQuery = false
		delete(normalized, UpstreamBalanceQueryExtraKey)
	}
	if hasQuery {
		payload, err := json.Marshal(rawQuery)
		if err != nil {
			return nil, ErrUpstreamBalanceQueryInvalid
		}
		var query UpstreamBalanceQuery
		if err := json.Unmarshal(payload, &query); err != nil {
			return nil, ErrUpstreamBalanceQueryInvalid
		}
		canonical, err := NormalizeUpstreamBalanceQuery(&query)
		if err != nil {
			return nil, err
		}
		normalized[UpstreamBalanceQueryExtraKey] = canonical
	}
	if enabled && !hasQuery {
		normalized[UpstreamBalanceQueryExtraKey] = &UpstreamBalanceQuery{Preset: UpstreamBalancePresetSub2API}
	}
	if !hasEnabled && !hasQuery {
		delete(normalized, UpstreamBalanceProbeEnabledExtraKey)
	}
	_ = platform
	return normalized, nil
}
