-- Replaces the flat strategies[]+params{} pair (one shared params bag for
-- every strategy in a candidate) with one instances[] array, each entry
-- carrying its own type/params/weight — lets a composite combine multiple
-- configured instances (including repeats of the same type, e.g. MA(20)
-- and MA(50) together) instead of collapsing them into one shared bag.
-- Existing rows are converted before the legacy columns are removed. The
-- old shared params object is copied into every generated instance; this is
-- the least-lossy representation possible for the former schema.
ALTER TABLE experiments ADD COLUMN IF NOT EXISTS instances TEXT NOT NULL DEFAULT '[]';

DO $$
BEGIN
	IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'experiments' AND column_name = 'strategies')
		AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'experiments' AND column_name = 'params') THEN
		UPDATE experiments AS experiment
		SET instances = (
			SELECT COALESCE(
				jsonb_agg(jsonb_build_object(
					'type', item.strategy_type,
					'params', CASE
						WHEN jsonb_typeof(COALESCE(NULLIF(experiment.params, ''), '{}')::jsonb) = 'object'
						THEN COALESCE(NULLIF(experiment.params, ''), '{}')::jsonb
						ELSE '{}'::jsonb
					END,
					'weight', 1
				)),
				'[]'::jsonb
			)::text
			FROM jsonb_array_elements_text(
				CASE
					WHEN jsonb_typeof(COALESCE(NULLIF(experiment.strategies, ''), '[]')::jsonb) = 'array'
					THEN COALESCE(NULLIF(experiment.strategies, ''), '[]')::jsonb
					ELSE '[]'::jsonb
				END
			) AS item(strategy_type)
		)
		WHERE experiment.instances = '[]';
	END IF;
END $$;

ALTER TABLE experiments DROP COLUMN IF EXISTS strategies;
ALTER TABLE experiments DROP COLUMN IF EXISTS params;
