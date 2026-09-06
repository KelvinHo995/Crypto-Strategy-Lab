-- Recover verifiable market provenance for rows created before migration 0013.
-- Unknown legacy values intentionally remain empty rather than being guessed.
UPDATE experiments AS e
SET
    pair = COALESCE(NULLIF(e.pair, ''), j.payload ->> 'pair', ''),
    timeframe = COALESCE(NULLIF(e.timeframe, ''), j.payload ->> 'timeframe', '')
FROM experiment_jobs AS j
WHERE j.id = e.id
  AND (e.pair = '' OR e.timeframe = '');

UPDATE experiments AS e
SET pair = t.pair
FROM (
    SELECT experiment_id, MIN(pair) AS pair
    FROM experiment_trades
    WHERE pair <> ''
    GROUP BY experiment_id
) AS t
WHERE t.experiment_id = e.id
  AND e.pair = '';
