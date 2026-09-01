-- Normalize rows written by the earlier pgx []byte-to-TEXT binding. The
-- migration is idempotent: only bytea-style hex literals are transformed.
UPDATE experiments
SET strategies = convert_from(decode(substr(strategies, 3), 'hex'), 'UTF8')
WHERE strategies LIKE '\x%';

UPDATE experiments
SET params = convert_from(decode(substr(params, 3), 'hex'), 'UTF8')
WHERE params LIKE '\x%';

UPDATE experiments
SET strategy_versions = convert_from(decode(substr(strategy_versions, 3), 'hex'), 'UTF8')
WHERE strategy_versions LIKE '\x%';
