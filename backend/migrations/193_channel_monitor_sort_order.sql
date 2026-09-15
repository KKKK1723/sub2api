-- 仅迁移时按名称初始化；后续改名、检测和重启均不重新计算顺序。
CREATE SEQUENCE IF NOT EXISTS channel_monitor_sort_order_seq;
ALTER TABLE channel_monitors ADD COLUMN IF NOT EXISTS sort_order BIGINT;

WITH labels AS (
    SELECT id, lower(coalesce(nullif(trim(group_name), ''), name)) AS label
    FROM channel_monitors
    WHERE sort_order IS NULL
), ranked AS (
    SELECT id, row_number() OVER (ORDER BY CASE
        WHEN label ~ '(openai|opnai).*官.*key[^0-9]*1([^0-9]|$)' THEN 10
        WHEN label ~ '(openai|opnai).*官.*key[^0-9]*2([^0-9]|$)' THEN 20
        WHEN label ~ '(满血.*pro|pro.*满血)' THEN 30
        WHEN label ~ '(混.*pro|pro.*混)' THEN 50
        WHEN label LIKE '%kiro%' THEN 80
        WHEN label LIKE '%pro%' THEN 40
        WHEN label LIKE '%plus%' THEN 60
        ELSE 70
    END, id) AS position
    FROM labels
)
UPDATE channel_monitors AS m SET sort_order = ranked.position
FROM ranked WHERE m.id = ranked.id;

SELECT setval('channel_monitor_sort_order_seq',
    greatest(coalesce(max(sort_order), 0) + 1,
        (SELECT last_value FROM channel_monitor_sort_order_seq)), false)
FROM channel_monitors;

ALTER TABLE channel_monitors ALTER COLUMN sort_order SET NOT NULL;
ALTER TABLE channel_monitors ALTER COLUMN sort_order SET DEFAULT nextval('channel_monitor_sort_order_seq');
ALTER SEQUENCE channel_monitor_sort_order_seq OWNED BY channel_monitors.sort_order;
CREATE INDEX IF NOT EXISTS idx_channel_monitors_sort_order ON channel_monitors (sort_order, id);
