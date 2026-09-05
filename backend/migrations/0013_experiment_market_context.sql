-- Market coordinates used by the backtest. Empty values preserve old rows as
-- unknown instead of assigning an unverified pair or timeframe.
ALTER TABLE experiments
    ADD COLUMN IF NOT EXISTS pair TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS timeframe TEXT NOT NULL DEFAULT '';
