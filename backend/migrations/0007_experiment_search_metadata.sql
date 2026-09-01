-- Keep every persisted candidate associated with its originating search.
-- The current API creates one candidate per search, but explicit metadata keeps
-- the schema compatible with future multi-candidate discovery runs.
ALTER TABLE experiments ADD COLUMN IF NOT EXISTS search_id TEXT;
ALTER TABLE experiments ADD COLUMN IF NOT EXISTS search_total INTEGER;

UPDATE experiments SET search_id = id
WHERE search_id IS NULL OR BTRIM(search_id) = '';
UPDATE experiments SET search_total = 1
WHERE search_total IS NULL OR search_total < 1;

ALTER TABLE experiments ALTER COLUMN search_id SET NOT NULL;
ALTER TABLE experiments ALTER COLUMN search_total SET NOT NULL;

CREATE INDEX IF NOT EXISTS experiments_search_id_idx
    ON experiments (search_id);
