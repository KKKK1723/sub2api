package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	UpstreamBalanceProbeEnabledExtraKey = "upstream_balance_probe_enabled"
	UpstreamBalanceQueryExtraKey        = "upstream_balance_query"
	UpstreamBalanceProbeExtraKey        = "upstream_balance_probe"

	UpstreamBalancePresetSub2API  = "sub2api"
	UpstreamBalancePresetCCSwitch = "ccswitch"
	UpstreamBalancePresetCustom   = "custom"
)

var (
	ErrUpstreamBalanceProbeUnavailable = infraerrors.ServiceUnavailable(
		"UPSTREAM_BALANCE_PROBE_UNAVAILABLE", "upstream balance probe is unavailable",
	)
	ErrUpstreamBalanceProbeAccountInvalid = infraerrors.BadRequest(
		"UPSTREAM_BALANCE_PROBE_ACCOUNT_INVALID", "account is not an API key account",
	)
	ErrUpstreamBalanceQueryInvalid = infraerrors.BadRequest(
		"UPSTREAM_BALANCE_QUERY_INVALID", "upstream balance query is invalid",
	)
	ErrUpstreamBalanceProbeIdentityChanged = infraerrors.Conflict(
		"UPSTREAM_BALANCE_PROBE_IDENTITY_CHANGED", "account identity or balance configuration changed during upstream balance probe; retry the probe",
	)
)

const (
	UpstreamBalanceProbeStatusOK          = "ok"
	UpstreamBalanceProbeStatusUnsupported = "unsupported"
	UpstreamBalanceProbeStatusFailed      = "failed"
)

// UpstreamBalanceQuery is a constrained alternative to CC Switch's JavaScript
// extractor. Only same-origin URLs and simple JSON field paths are accepted.
type UpstreamBalanceQuery struct {
	Preset      string `json:"preset"`
	Method      string `json:"method,omitempty"`
	URL         string `json:"url,omitempty"`
	AuthType    string `json:"auth_type,omitempty"`
	BalancePath string `json:"balance_path,omitempty"`
	UnitPath    string `json:"unit_path,omitempty"`
	PlanPath    string `json:"plan_path,omitempty"`
	TotalPath   string `json:"total_path,omitempty"`
	UsedPath    string `json:"used_path,omitempty"`
	ValidPath   string `json:"valid_path,omitempty"`
	ErrorPath   string `json:"error_path,omitempty"`
	Unit        string `json:"unit,omitempty"`
}

type upstreamBalanceResolvedQuery struct {
	Method       string
	URLTemplate  string
	AuthType     string
	BalancePaths []string
	UnitPaths    []string
	PlanPaths    []string
	TotalPaths   []string
	UsedPaths    []string
	ValidPaths   []string
	ErrorPaths   []string
	DefaultUnit  string
	UserAgent    string
}

type UpstreamBalanceData struct {
	Balance  float64  `json:"balance"`
	Unit     string   `json:"unit"`
	PlanName string   `json:"plan_name,omitempty"`
	Total    *float64 `json:"total,omitempty"`
	Used     *float64 `json:"used,omitempty"`
	IsValid  *bool    `json:"is_valid,omitempty"`
}

type UpstreamBalanceProbeSnapshot struct {
	Status        string               `json:"status"`
	Data          *UpstreamBalanceData `json:"data,omitempty"`
	ReceivedAt    *time.Time           `json:"received_at,omitempty"`
	FreshUntil    *time.Time           `json:"fresh_until,omitempty"`
	LastAttemptAt time.Time            `json:"last_attempt_at"`
	NextProbeAt   time.Time            `json:"next_probe_at"`
	FailureCount  int                  `json:"failure_count,omitempty"`
	HTTPStatus    int                  `json:"http_status,omitempty"`
	LastError     string               `json:"last_error,omitempty"`
}

type UpstreamBalanceProbeResult struct {
	AccountID int64                         `json:"account_id"`
	Snapshot  *UpstreamBalanceProbeSnapshot `json:"snapshot,omitempty"`
	Error     string                        `json:"error,omitempty"`
}

var upstreamBalancePathSegmentPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// NormalizeUpstreamBalanceQuery validates and canonicalizes an account setting.
func NormalizeUpstreamBalanceQuery(query *UpstreamBalanceQuery) (*UpstreamBalanceQuery, error) {
	if query == nil {
		return nil, ErrUpstreamBalanceQueryInvalid
	}
	normalized := *query
	normalized.Preset = strings.ToLower(strings.TrimSpace(normalized.Preset))
	if normalized.Preset == "" {
		normalized.Preset = UpstreamBalancePresetSub2API
	}
	switch normalized.Preset {
	case UpstreamBalancePresetSub2API, UpstreamBalancePresetCCSwitch:
		return &UpstreamBalanceQuery{Preset: normalized.Preset}, nil
	case UpstreamBalancePresetCustom:
	default:
		return nil, invalidUpstreamBalanceQuery("preset must be sub2api, ccswitch, or custom")
	}

	normalized.Method = strings.ToUpper(strings.TrimSpace(normalized.Method))
	if normalized.Method == "" {
		normalized.Method = http.MethodGet
	}
	if normalized.Method != http.MethodGet && normalized.Method != http.MethodPost {
		return nil, invalidUpstreamBalanceQuery("method must be GET or POST")
	}
	normalized.URL = strings.TrimSpace(normalized.URL)
	if normalized.URL == "" || len(normalized.URL) > 2048 {
		return nil, invalidUpstreamBalanceQuery("url is required and must not exceed 2048 characters")
	}
	if strings.Contains(normalized.URL, "{{apiKey}}") || strings.Contains(normalized.URL, "{{api_key}}") {
		return nil, invalidUpstreamBalanceQuery("API keys are not allowed in the URL")
	}
	placeholderCheck := strings.NewReplacer(
		"{{baseUrl}}", "https://balance.invalid/v1",
		"{{origin}}", "https://balance.invalid",
	).Replace(normalized.URL)
	if strings.Contains(placeholderCheck, "{{") || strings.Contains(placeholderCheck, "}}") {
		return nil, invalidUpstreamBalanceQuery("url contains an unsupported template variable")
	}
	parsed, err := url.Parse(placeholderCheck)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, invalidUpstreamBalanceQuery("url must be absolute or use {{baseUrl}}/{{origin}}")
	}

	normalized.AuthType = strings.ToLower(strings.TrimSpace(normalized.AuthType))
	if normalized.AuthType == "" {
		normalized.AuthType = "bearer"
	}
	if normalized.AuthType != "bearer" && normalized.AuthType != "x-api-key" {
		return nil, invalidUpstreamBalanceQuery("auth_type must be bearer or x-api-key")
	}
	paths := []struct {
		name  string
		value *string
	}{
		{"balance_path", &normalized.BalancePath},
		{"unit_path", &normalized.UnitPath},
		{"plan_path", &normalized.PlanPath},
		{"total_path", &normalized.TotalPath},
		{"used_path", &normalized.UsedPath},
		{"valid_path", &normalized.ValidPath},
		{"error_path", &normalized.ErrorPath},
	}
	for _, field := range paths {
		*field.value = strings.TrimSpace(*field.value)
		if field.name == "balance_path" && *field.value == "" {
			return nil, invalidUpstreamBalanceQuery("balance_path is required")
		}
		if *field.value != "" && !validUpstreamBalancePath(*field.value) {
			return nil, invalidUpstreamBalanceQuery(field.name + " is not a valid JSON field path")
		}
	}
	normalized.Unit = strings.ToUpper(strings.TrimSpace(normalized.Unit))
	if len(normalized.Unit) > 16 {
		return nil, invalidUpstreamBalanceQuery("unit must not exceed 16 characters")
	}
	return &normalized, nil
}

func invalidUpstreamBalanceQuery(message string) error {
	return infraerrors.BadRequest("UPSTREAM_BALANCE_QUERY_INVALID", message)
}

func validUpstreamBalancePath(path string) bool {
	if len(path) == 0 || len(path) > 160 {
		return false
	}
	segments := strings.Split(path, ".")
	if len(segments) > 16 {
		return false
	}
	for _, segment := range segments {
		if !upstreamBalancePathSegmentPattern.MatchString(segment) {
			return false
		}
	}
	return true
}

func resolveUpstreamBalanceQuery(query *UpstreamBalanceQuery) (*upstreamBalanceResolvedQuery, error) {
	normalized, err := NormalizeUpstreamBalanceQuery(query)
	if err != nil {
		return nil, err
	}
	switch normalized.Preset {
	case UpstreamBalancePresetSub2API, UpstreamBalancePresetCCSwitch:
		return &upstreamBalanceResolvedQuery{
			Method:       http.MethodGet,
			URLTemplate:  "{{baseUrl}}/v1/usage",
			AuthType:     "bearer",
			BalancePaths: []string{"remaining", "quota.remaining", "balance", "data.remaining", "data.balance"},
			UnitPaths:    []string{"unit", "quota.unit", "data.unit"},
			PlanPaths:    []string{"planName", "data.planName"},
			TotalPaths:   []string{"quota.limit", "total", "data.total"},
			UsedPaths:    []string{"quota.used", "used", "data.used"},
			ValidPaths:   []string{"is_active", "isValid", "data.is_active", "data.isValid"},
			ErrorPaths:   []string{"error", "data.error"},
			DefaultUnit:  "USD",
			UserAgent:    "sub2api-upstream-balance/1.0",
		}, nil
	default:
		return &upstreamBalanceResolvedQuery{
			Method: normalized.Method, URLTemplate: normalized.URL, AuthType: normalized.AuthType,
			BalancePaths: compactBalancePaths(normalized.BalancePath),
			UnitPaths:    compactBalancePaths(normalized.UnitPath), PlanPaths: compactBalancePaths(normalized.PlanPath),
			TotalPaths: compactBalancePaths(normalized.TotalPath), UsedPaths: compactBalancePaths(normalized.UsedPath),
			ValidPaths: compactBalancePaths(normalized.ValidPath), ErrorPaths: compactBalancePaths(normalized.ErrorPath),
			DefaultUnit: normalized.Unit, UserAgent: "sub2api-upstream-balance/1.0",
		}, nil
	}
}

func compactBalancePaths(paths ...string) []string {
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		if path = strings.TrimSpace(path); path != "" {
			out = append(out, path)
		}
	}
	return out
}

func decodeUpstreamBalanceQuery(extra map[string]any) *UpstreamBalanceQuery {
	if extra == nil || extra[UpstreamBalanceQueryExtraKey] == nil {
		return nil
	}
	payload, err := json.Marshal(extra[UpstreamBalanceQueryExtraKey])
	if err != nil {
		return nil
	}
	var query UpstreamBalanceQuery
	if err := json.Unmarshal(payload, &query); err != nil {
		return nil
	}
	normalized, err := NormalizeUpstreamBalanceQuery(&query)
	if err != nil {
		return nil
	}
	return normalized
}

func decodeUpstreamBalanceProbeSnapshot(extra map[string]any) *UpstreamBalanceProbeSnapshot {
	if extra == nil || extra[UpstreamBalanceProbeExtraKey] == nil {
		return nil
	}
	payload, err := json.Marshal(extra[UpstreamBalanceProbeExtraKey])
	if err != nil {
		return nil
	}
	var snapshot UpstreamBalanceProbeSnapshot
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		return nil
	}
	switch snapshot.Status {
	case UpstreamBalanceProbeStatusOK, UpstreamBalanceProbeStatusUnsupported, UpstreamBalanceProbeStatusFailed:
		return &snapshot
	default:
		return nil
	}
}

func upstreamBalanceProbeEnabled(account *Account) bool {
	if account == nil || account.Extra == nil {
		return false
	}
	enabled, ok := account.Extra[UpstreamBalanceProbeEnabledExtraKey].(bool)
	return ok && enabled
}

func isUpstreamBalanceProbeAccount(account *Account) bool {
	return account != nil && account.Type == AccountTypeAPIKey
}

func upstreamBalanceProbeIdentity(account *Account) map[string]any {
	if account == nil {
		return nil
	}
	identity := map[string]any{
		"platform":    account.Platform,
		"type":        account.Type,
		"credentials": account.Credentials,
		"proxy_id":    nil,
		"query":       nil,
	}
	if account.ProxyID != nil {
		identity["proxy_id"] = *account.ProxyID
	}
	if account.Extra != nil {
		identity["query"] = account.Extra[UpstreamBalanceQueryExtraKey]
	}
	return identity
}

func resolveUpstreamBalanceURL(baseURL, template string) (string, error) {
	base, err := url.Parse(strings.TrimRight(strings.TrimSpace(baseURL), "/"))
	if err != nil || !base.IsAbs() || base.Host == "" {
		return "", fmt.Errorf("invalid base URL")
	}
	if template == "{{baseUrl}}/v1/usage" {
		return buildOpenAIEndpointURL(baseURL, "/v1/usage"), nil
	}
	origin := base.Scheme + "://" + base.Host
	resolved := strings.NewReplacer("{{baseUrl}}", strings.TrimRight(base.String(), "/"), "{{origin}}", origin).Replace(strings.TrimSpace(template))
	parsed, err := url.Parse(resolved)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" {
		return "", fmt.Errorf("invalid query URL")
	}
	return parsed.String(), nil
}

func sameUpstreamBalanceOrigin(left, right string) bool {
	leftURL, leftErr := url.Parse(left)
	rightURL, rightErr := url.Parse(right)
	return leftErr == nil && rightErr == nil && strings.EqualFold(leftURL.Scheme, rightURL.Scheme) && strings.EqualFold(leftURL.Host, rightURL.Host)
}

func parseUpstreamBalanceResponse(body []byte, query *upstreamBalanceResolvedQuery) (*UpstreamBalanceData, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var root any
	if err := decoder.Decode(&root); err != nil {
		return nil, err
	}
	if value, ok := firstUpstreamBalancePath(root, query.ErrorPaths); ok && upstreamBalanceErrorPresent(value) {
		return nil, fmt.Errorf("upstream reported an error")
	}
	balanceValue, ok := firstUpstreamBalancePath(root, query.BalancePaths)
	if !ok {
		return nil, fmt.Errorf("balance field is missing")
	}
	balance, ok := upstreamBalanceNumber(balanceValue)
	if !ok {
		return nil, fmt.Errorf("balance field is not numeric")
	}
	unit := strings.TrimSpace(query.DefaultUnit)
	if value, ok := firstUpstreamBalancePath(root, query.UnitPaths); ok {
		if parsed, valid := upstreamBalanceString(value); valid {
			unit = parsed
		}
	}
	if unit == "" {
		unit = "USD"
	}
	if len(unit) > 16 {
		return nil, fmt.Errorf("unit is too long")
	}
	data := &UpstreamBalanceData{Balance: balance, Unit: unit}
	if value, ok := firstUpstreamBalancePath(root, query.PlanPaths); ok {
		if parsed, valid := upstreamBalanceString(value); valid && len(parsed) <= 128 {
			data.PlanName = parsed
		}
	}
	if value, ok := firstUpstreamBalancePath(root, query.TotalPaths); ok {
		if parsed, valid := upstreamBalanceNumber(value); valid {
			data.Total = &parsed
		}
	}
	if value, ok := firstUpstreamBalancePath(root, query.UsedPaths); ok {
		if parsed, valid := upstreamBalanceNumber(value); valid {
			data.Used = &parsed
		}
	}
	if value, ok := firstUpstreamBalancePath(root, query.ValidPaths); ok {
		if parsed, valid := upstreamBalanceBool(value); valid {
			data.IsValid = &parsed
		}
	}
	return data, nil
}

func firstUpstreamBalancePath(root any, paths []string) (any, bool) {
	for _, path := range paths {
		if value, ok := upstreamBalancePath(root, path); ok && value != nil {
			return value, true
		}
	}
	return nil, false
}

func upstreamBalancePath(root any, path string) (any, bool) {
	current := root
	for _, segment := range strings.Split(path, ".") {
		switch typed := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = typed[segment]
			if !ok {
				return nil, false
			}
		case []any:
			index, err := strconv.Atoi(segment)
			if err != nil || index < 0 || index >= len(typed) {
				return nil, false
			}
			current = typed[index]
		default:
			return nil, false
		}
	}
	return current, true
}

func upstreamBalanceNumber(value any) (float64, bool) {
	var parsed float64
	var err error
	switch typed := value.(type) {
	case json.Number:
		parsed, err = typed.Float64()
	case float64:
		parsed = typed
	case float32:
		parsed = float64(typed)
	case int:
		parsed = float64(typed)
	case int64:
		parsed = float64(typed)
	case string:
		parsed, err = strconv.ParseFloat(strings.TrimSpace(typed), 64)
	default:
		return 0, false
	}
	return parsed, err == nil && !math.IsNaN(parsed) && !math.IsInf(parsed, 0)
}

func upstreamBalanceString(value any) (string, bool) {
	parsed, ok := value.(string)
	parsed = strings.TrimSpace(parsed)
	return parsed, ok && parsed != ""
}

func upstreamBalanceBool(value any) (bool, bool) {
	switch typed := value.(type) {
	case bool:
		return typed, true
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(typed))
		return parsed, err == nil
	case json.Number:
		parsed, err := typed.Float64()
		return parsed != 0, err == nil
	default:
		return false, false
	}
}

func upstreamBalanceErrorPresent(value any) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case bool:
		return typed
	case string:
		return strings.TrimSpace(typed) != ""
	case json.Number:
		parsed, err := typed.Float64()
		return err != nil || parsed != 0
	case []any:
		return len(typed) > 0
	case map[string]any:
		return len(typed) > 0
	default:
		return true
	}
}
