ALTER TABLE announcements
    ADD COLUMN IF NOT EXISTS pinned BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_announcements_pinned ON announcements(pinned);
