ALTER TABLE experiments ADD COLUMN IF NOT EXISTS updated_at BIGINT;
UPDATE experiments SET updated_at = created_at WHERE updated_at IS NULL;
ALTER TABLE experiments ALTER COLUMN updated_at SET NOT NULL;

CREATE INDEX IF NOT EXISTS experiments_running_updated_at_idx
    ON experiments (updated_at) WHERE status = 'RUNNING';
