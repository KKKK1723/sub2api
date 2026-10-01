//go:build unit

package service

import "testing"

func TestAggregateConfiguredProbeStatusesUsesGreenYellowRedPriority(t *testing.T) {
	fast := 120
	slow := 240
	got := AggregateConfiguredProbeStatuses([]MonitorProbe{
		{Enabled: true, Status: MonitorStatusFailed},
		{Enabled: true, Status: MonitorStatusOperational, LatencyMs: &slow},
		{Enabled: true, Status: MonitorStatusOperational, LatencyMs: &fast},
		{Enabled: false, Status: MonitorStatusFailed},
	})
	if got == nil || got.Status != MonitorStatusOperational || got.LatencyMs == nil || *got.LatencyMs != fast {
		t.Fatalf("expected fastest green enabled probe, got %#v", got)
	}

	got = AggregateConfiguredProbeStatuses([]MonitorProbe{
		{Enabled: true, Status: MonitorStatusFailed},
		{Enabled: true, Status: MonitorStatusDegraded},
	})
	if got == nil || got.Status != MonitorStatusDegraded {
		t.Fatalf("expected yellow when no green probe exists, got %#v", got)
	}
}

func TestBuildStatusSummaryUsesHealthyProbeAcrossModels(t *testing.T) {
	fast := 120
	slow := 240
	got := buildStatusSummary(
		map[string]*ChannelMonitorLatest{
			"gpt-pro-primary": {Model: "gpt-pro-primary", Status: "failed", LatencyMs: nil},
			"gpt-pro-backup":  {Model: "gpt-pro-backup", Status: "operational", LatencyMs: &fast},
			"gpt-pro-third":   {Model: "gpt-pro-third", Status: "operational", LatencyMs: &slow},
		},
		nil,
		"gpt-pro-primary",
		nil,
		[]string{"gpt-pro-primary", "gpt-pro-backup", "gpt-pro-third"},
	)

	if got.PrimaryStatus != "operational" {
		t.Fatalf("expected healthy probe to determine group status, got %q", got.PrimaryStatus)
	}
	if got.PrimaryLatencyMs == nil || *got.PrimaryLatencyMs != fast {
		t.Fatalf("expected fastest healthy probe latency %d, got %v", fast, got.PrimaryLatencyMs)
	}
}

func TestBuildStatusSummaryKeepsPrimaryWhenNoProbeModelSummary(t *testing.T) {
	latency := 180
	got := buildStatusSummary(
		map[string]*ChannelMonitorLatest{
			"gpt-pro-primary": {Model: "gpt-pro-primary", Status: "degraded", LatencyMs: &latency},
		},
		nil,
		"gpt-pro-primary",
		nil,
		nil,
	)

	if got.PrimaryStatus != "degraded" {
		t.Fatalf("expected primary status fallback, got %q", got.PrimaryStatus)
	}
}
