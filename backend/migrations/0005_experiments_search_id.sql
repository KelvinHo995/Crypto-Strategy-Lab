ALTER TABLE experiments ADD COLUMN IF NOT EXISTS search_id TEXT;
ALTER TABLE experiments ADD COLUMN IF NOT EXISTS search_total INTEGER;
UPDATE experiments SET search_id = id WHERE search_id IS NULL;
UPDATE experiments SET search_total = 1 WHERE search_total IS NULL;
ALTER TABLE experiments ALTER COLUMN search_id SET NOT NULL;
ALTER TABLE experiments ALTER COLUMN search_total SET NOT NULL;

CREATE INDEX IF NOT EXISTS experiments_search_id_idx ON experiments (search_id);
