-- 为监控历史增加探针维度，确保同模型的多个探针可以独立参与状态聚合。
ALTER TABLE channel_monitor_histories
    ADD COLUMN IF NOT EXISTS probe_index INT NOT NULL DEFAULT -1;

CREATE INDEX IF NOT EXISTS idx_channel_monitor_histories_monitor_probe_checked
    ON channel_monitor_histories (monitor_id, probe_index, checked_at DESC);
