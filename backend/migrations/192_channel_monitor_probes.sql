-- Migration: 192_channel_monitor_probes
-- Stores multiple encrypted endpoint/API key probes per monitor.
ALTER TABLE channel_monitors
    ADD COLUMN IF NOT EXISTS probes JSONB NOT NULL DEFAULT '[]'::jsonb;
