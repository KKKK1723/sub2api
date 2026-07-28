package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"golang.org/x/sync/errgroup"
)

const upstreamBalanceProbeLeaderLockKey = "upstream:balance:probe:leader"

type upstreamBalanceProbeSnapshotWriter interface {
	UpdateUpstreamBalanceProbeSnapshot(context.Context, *Account, *UpstreamBalanceProbeSnapshot) error
}

type upstreamBalanceProbeDueAccountLister interface {
	ListDueUpstreamBalanceProbeAccounts(context.Context, time.Time, int) ([]Account, error)
}

// RunBalanceDue executes one bounded periodic balance-probe batch. It shares
// the existing billing-probe cadence, settings, concurrency, and HTTP limits.
func (s *UpstreamBillingProbeService) RunBalanceDue(ctx context.Context) error {
	if s == nil || s.accountRepo == nil {
		return nil
	}
	s.cycleMu.Lock()
	defer s.cycleMu.Unlock()
	settings, err := s.getSettings(ctx)
	if err != nil {
		return err
	}
	if !settings.Enabled {
		return nil
	}
	release, acquired, err := s.tryAcquireLeaderLock(ctx, upstreamBalanceProbeLeaderLockKey)
	if err != nil {
		return fmt.Errorf("acquire upstream balance probe leader lock: %w", err)
	}
	if !acquired {
		return nil
	}
	defer release()

	lockNow := time.Now()
	cadenceKey := fmt.Sprintf("%s:%d", upstreamBalanceProbeLeaderLockKey, lockNow.Unix()/int64(upstreamBillingProbeCycleInterval/time.Second))
	cadenceRelease, acquired, err := s.tryAcquireLeaderLock(ctx, cadenceKey)
	if err != nil {
		return fmt.Errorf("acquire upstream balance probe cadence lock: %w", err)
	}
	if !acquired {
		return nil
	}
	defer releaseUpstreamBillingProbeLeaderLock(cadenceRelease, lockNow.Truncate(upstreamBillingProbeCycleInterval).Add(upstreamBillingProbeCycleInterval))

	now := s.currentTime()
	accounts, err := s.listDueUpstreamBalanceAccounts(ctx, now)
	if err != nil {
		return fmt.Errorf("list enabled upstream balance probes: %w", err)
	}
	due := make([]Account, 0, len(accounts))
	for i := range accounts {
		account := accounts[i]
		if !isUpstreamBalanceProbeAccount(&account) || !account.IsActive() || !upstreamBalanceProbeEnabled(&account) || decodeUpstreamBalanceQuery(account.Extra) == nil {
			continue
		}
		snapshot := decodeUpstreamBalanceProbeSnapshot(account.Extra)
		if snapshot != nil && !snapshot.NextProbeAt.IsZero() && now.Before(snapshot.NextProbeAt) {
			continue
		}
		due = append(due, account)
	}
	sort.SliceStable(due, func(i, j int) bool {
		left := upstreamBalanceNextProbeAt(&due[i])
		right := upstreamBalanceNextProbeAt(&due[j])
		if left.IsZero() != right.IsZero() {
			return left.IsZero()
		}
		if left.Equal(right) {
			return due[i].ID < due[j].ID
		}
		return left.Before(right)
	})
	if len(due) > upstreamBillingProbeMaxPerCycle {
		due = due[:upstreamBillingProbeMaxPerCycle]
	}
	var group errgroup.Group
	for i := range due {
		accountID := due[i].ID
		group.Go(func() error {
			if _, probeErr := s.probeUpstreamBalanceAccount(ctx, accountID, settings.IntervalMinutes, true); probeErr != nil {
				logger.LegacyPrintf("service.upstream_balance_probe", "probe_due_failed: account_id=%d err=%v", accountID, probeErr)
			}
			return nil
		})
	}
	return group.Wait()
}

func upstreamBalanceNextProbeAt(account *Account) time.Time {
	if snapshot := decodeUpstreamBalanceProbeSnapshot(account.Extra); snapshot != nil {
		return snapshot.NextProbeAt
	}
	return time.Time{}
}

func (s *UpstreamBillingProbeService) listDueUpstreamBalanceAccounts(ctx context.Context, now time.Time) ([]Account, error) {
	if lister, ok := s.accountRepo.(upstreamBalanceProbeDueAccountLister); ok {
		return lister.ListDueUpstreamBalanceProbeAccounts(ctx, now, upstreamBillingProbeMaxPerCycle)
	}
	return s.accountRepo.FindByExtraField(ctx, UpstreamBalanceProbeEnabledExtraKey, true)
}

func (s *UpstreamBillingProbeService) ProbeUpstreamBalance(ctx context.Context, accountID int64) (*UpstreamBalanceProbeSnapshot, error) {
	if s == nil || s.accountRepo == nil {
		return nil, ErrUpstreamBalanceProbeUnavailable
	}
	settings, err := s.getSettings(ctx)
	if err != nil {
		return nil, err
	}
	return s.probeUpstreamBalanceAccount(ctx, accountID, settings.IntervalMinutes, false)
}

func (s *UpstreamBillingProbeService) ProbeUpstreamBalances(ctx context.Context, accountIDs []int64) []UpstreamBalanceProbeResult {
	if len(accountIDs) > upstreamBillingProbeMaxPerCycle {
		accountIDs = accountIDs[:upstreamBillingProbeMaxPerCycle]
	}
	results := make([]UpstreamBalanceProbeResult, len(accountIDs))
	if s == nil || s.accountRepo == nil {
		for i, accountID := range accountIDs {
			results[i] = UpstreamBalanceProbeResult{AccountID: accountID, Error: ErrUpstreamBalanceProbeUnavailable.Error()}
		}
		return results
	}
	settings, settingsErr := s.getSettings(ctx)
	var group errgroup.Group
	for i, accountID := range accountIDs {
		i, accountID := i, accountID
		results[i].AccountID = accountID
		group.Go(func() error {
			if settingsErr != nil {
				results[i].Error = safeUpstreamBalanceProbeError(settingsErr)
				return nil
			}
			snapshot, err := s.probeUpstreamBalanceAccount(ctx, accountID, settings.IntervalMinutes, false)
			if err != nil {
				results[i].Error = safeUpstreamBalanceProbeError(err)
				return nil
			}
			results[i].Snapshot = snapshot
			return nil
		})
	}
	_ = group.Wait()
	return results
}

func (s *UpstreamBillingProbeService) probeUpstreamBalanceAccount(ctx context.Context, accountID int64, intervalMinutes int, requireEnabled bool) (*UpstreamBalanceProbeSnapshot, error) {
	key := "balance:" + strconv.FormatInt(accountID, 10)
	value, err, _ := s.probeGroup.Do(key, func() (any, error) {
		select {
		case s.probeSlots <- struct{}{}:
			defer func() { <-s.probeSlots }()
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		account, err := s.accountRepo.GetByID(ctx, accountID)
		if err != nil {
			return nil, err
		}
		if !isUpstreamBalanceProbeAccount(account) {
			return nil, ErrUpstreamBalanceProbeAccountInvalid
		}
		if decodeUpstreamBalanceQuery(account.Extra) == nil {
			return nil, ErrUpstreamBalanceQueryInvalid
		}
		if requireEnabled {
			if !account.IsActive() || !upstreamBalanceProbeEnabled(account) {
				return nil, nil
			}
			if snapshot := decodeUpstreamBalanceProbeSnapshot(account.Extra); snapshot != nil && !snapshot.NextProbeAt.IsZero() && s.currentTime().Before(snapshot.NextProbeAt) {
				return nil, nil
			}
		}
		return s.probeLoadedUpstreamBalanceAccount(ctx, account, intervalMinutes)
	})
	if err != nil || value == nil {
		return nil, err
	}
	snapshot, ok := value.(*UpstreamBalanceProbeSnapshot)
	if !ok {
		return nil, fmt.Errorf("invalid upstream balance probe result")
	}
	return snapshot, nil
}

func (s *UpstreamBillingProbeService) probeLoadedUpstreamBalanceAccount(ctx context.Context, account *Account, intervalMinutes int) (*UpstreamBalanceProbeSnapshot, error) {
	now := s.currentTime().UTC()
	if s.accountTestService == nil || s.accountTestService.httpUpstream == nil {
		return s.persistUpstreamBalanceFailure(ctx, account, intervalMinutes, now, 0, "transport_unavailable", 0)
	}
	apiKey := strings.TrimSpace(account.GetCredential("api_key"))
	if apiKey == "" {
		return s.persistUpstreamBalanceFailure(ctx, account, intervalMinutes, now, 0, "missing_api_key", 0)
	}
	baseURL := strings.TrimSpace(account.GetCredential("base_url"))
	if baseURL == "" {
		baseURL = defaultUpstreamBalanceBaseURL(account.Platform)
	}
	baseURL, err := s.accountTestService.validateUpstreamBaseURL(baseURL)
	if err != nil {
		return s.persistUpstreamBalanceFailure(ctx, account, intervalMinutes, now, 0, "invalid_base_url", 0)
	}
	query, err := resolveUpstreamBalanceQuery(decodeUpstreamBalanceQuery(account.Extra))
	if err != nil {
		return nil, err
	}
	probeURL, err := resolveUpstreamBalanceURL(baseURL, query.URLTemplate)
	if err != nil {
		return s.persistUpstreamBalanceFailure(ctx, account, intervalMinutes, now, 0, "invalid_query_url", 0)
	}
	validatedProbeURL, err := s.accountTestService.validateUpstreamBaseURL(probeURL)
	if err != nil || !sameUpstreamBalanceOrigin(baseURL, validatedProbeURL) {
		return s.persistUpstreamBalanceFailure(ctx, account, intervalMinutes, now, 0, "invalid_query_url", 0)
	}
	proxyURL := ""
	if account.ProxyID != nil {
		if account.Proxy == nil {
			return s.persistUpstreamBalanceFailure(ctx, account, intervalMinutes, now, 0, "proxy_unavailable", 0)
		}
		if account.Proxy.ID != *account.ProxyID {
			return nil, ErrUpstreamBalanceProbeIdentityChanged
		}
		proxyURL = account.Proxy.URL()
	}

	probeCtx, cancel := context.WithTimeout(ctx, upstreamBillingProbeRequestTimeout)
	defer cancel()
	var requestBody io.Reader
	if query.Method == http.MethodPost {
		requestBody = bytes.NewReader(nil)
	}
	req, err := http.NewRequestWithContext(probeCtx, query.Method, validatedProbeURL, requestBody)
	if err != nil {
		return s.persistUpstreamBalanceFailure(ctx, account, intervalMinutes, now, 0, "request_build_failed", 0)
	}
	req = req.WithContext(WithHTTPUpstreamRedirectsDisabled(req.Context()))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", query.UserAgent)
	if query.Method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	if query.AuthType == "x-api-key" {
		req.Header.Set("X-API-Key", apiKey)
	} else {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	account.ApplyHeaderOverrides(req.Header)
	var tlsProfile *tlsfingerprint.Profile
	if s.accountTestService.tlsFPProfileService != nil {
		tlsProfile = s.accountTestService.tlsFPProfileService.ResolveTLSProfile(account)
	}
	resp, err := s.accountTestService.httpUpstream.DoWithTLS(req, proxyURL, account.ID, account.Concurrency, tlsProfile)
	if err != nil {
		return s.persistUpstreamBalanceFailure(ctx, account, intervalMinutes, now, 0, "request_failed", 0)
	}
	if resp == nil || resp.Body == nil {
		return s.persistUpstreamBalanceFailure(ctx, account, intervalMinutes, now, 0, "empty_response", 0)
	}
	defer func() { _ = resp.Body.Close() }()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, upstreamBillingProbeMaxBodyBytes+1))
	if readErr != nil {
		return s.persistUpstreamBalanceFailure(ctx, account, intervalMinutes, now, resp.StatusCode, "response_read_failed", retryAfter(resp.Header, now))
	}
	if len(body) > upstreamBillingProbeMaxBodyBytes {
		return s.persistUpstreamBalanceFailure(ctx, account, intervalMinutes, now, resp.StatusCode, "response_too_large", retryAfter(resp.Header, now))
	}
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed {
		return s.persistUpstreamBalanceFailure(ctx, account, intervalMinutes, now, resp.StatusCode, "unsupported", retryAfter(resp.Header, now))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return s.persistUpstreamBalanceFailure(ctx, account, intervalMinutes, now, resp.StatusCode, "http_error", retryAfter(resp.Header, now))
	}
	data, err := parseUpstreamBalanceResponse(body, query)
	if err != nil {
		return s.persistUpstreamBalanceFailure(ctx, account, intervalMinutes, now, resp.StatusCode, "invalid_response", retryAfter(resp.Header, now))
	}
	snapshot := &UpstreamBalanceProbeSnapshot{
		Status: UpstreamBalanceProbeStatusOK, Data: data,
		ReceivedAt: probeTimePtr(now), FreshUntil: probeTimePtr(now.Add(2 * time.Duration(intervalMinutes) * time.Minute)),
		LastAttemptAt: now, NextProbeAt: now.Add(nextProbeDelay(intervalMinutes, 0)), HTTPStatus: resp.StatusCode,
	}
	if err := s.updateUpstreamBalanceSnapshot(ctx, account, snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func defaultUpstreamBalanceBaseURL(platform string) string {
	switch platform {
	case PlatformOpenAI:
		return "https://api.openai.com"
	case PlatformAnthropic:
		return "https://api.anthropic.com"
	case PlatformGemini:
		return "https://generativelanguage.googleapis.com"
	case PlatformGrok:
		return "https://api.x.ai/v1"
	default:
		return ""
	}
}

func (s *UpstreamBillingProbeService) persistUpstreamBalanceFailure(ctx context.Context, account *Account, intervalMinutes int, now time.Time, statusCode int, reason string, retryAfterDuration time.Duration) (*UpstreamBalanceProbeSnapshot, error) {
	previous := decodeUpstreamBalanceProbeSnapshot(account.Extra)
	failureCount := 1
	if previous != nil {
		failureCount = previous.FailureCount + 1
	}
	status := UpstreamBalanceProbeStatusFailed
	if reason == "unsupported" {
		status = UpstreamBalanceProbeStatusUnsupported
	}
	snapshot := &UpstreamBalanceProbeSnapshot{
		Status: status, LastAttemptAt: now, NextProbeAt: now.Add(nextProbeDelay(intervalMinutes, retryAfterDuration)),
		FailureCount: failureCount, HTTPStatus: statusCode, LastError: reason,
	}
	if previous != nil {
		snapshot.Data, snapshot.ReceivedAt, snapshot.FreshUntil = previous.Data, previous.ReceivedAt, previous.FreshUntil
		if snapshot.FreshUntil == nil && previous.Status == UpstreamBalanceProbeStatusOK && previous.ReceivedAt != nil {
			snapshot.FreshUntil = probeTimePtr(previous.ReceivedAt.Add(2 * time.Duration(intervalMinutes) * time.Minute))
		}
	}
	if err := s.updateUpstreamBalanceSnapshot(ctx, account, snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (s *UpstreamBillingProbeService) updateUpstreamBalanceSnapshot(ctx context.Context, account *Account, snapshot *UpstreamBalanceProbeSnapshot) error {
	writer, ok := s.accountRepo.(upstreamBalanceProbeSnapshotWriter)
	if !ok {
		return ErrUpstreamBalanceProbeUnavailable
	}
	return writer.UpdateUpstreamBalanceProbeSnapshot(ctx, account, snapshot)
}

func safeUpstreamBalanceProbeError(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrUpstreamBalanceProbeAccountInvalid), errors.Is(err, ErrUpstreamBalanceQueryInvalid):
		return err.Error()
	case errors.Is(err, ErrUpstreamBalanceProbeUnavailable):
		return ErrUpstreamBalanceProbeUnavailable.Error()
	default:
		return "probe_failed"
	}
}
