-- Normalize rows written by the earlier pgx []byte-to-TEXT binding. The
-- migration is idempotent: only bytea-style hex literals are transformed.
-- The column-existence guards keep this replayable even after migration
-- 0008 dropped strategies/params in favor of one instances column — the
-- runner has no applied-migrations tracking, so every file must stay safe
-- to run against whatever schema state actually exists, not just the one
-- it was originally written against.
DO $$
BEGIN
	IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'experiments' AND column_name = 'strategies') THEN
		UPDATE experiments
		SET strategies = convert_from(decode(substr(strategies, 3), 'hex'), 'UTF8')
		WHERE strategies LIKE '\x%';
	END IF;

	IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'experiments' AND column_name = 'params') THEN
		UPDATE experiments
		SET params = convert_from(decode(substr(params, 3), 'hex'), 'UTF8')
		WHERE params LIKE '\x%';
	END IF;

	IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'experiments' AND column_name = 'strategy_versions') THEN
		UPDATE experiments
		SET strategy_versions = convert_from(decode(substr(strategy_versions, 3), 'hex'), 'UTF8')
		WHERE strategy_versions LIKE '\x%';
	END IF;
END $$;
