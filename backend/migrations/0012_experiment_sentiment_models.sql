-- Runtime model identities consumed by SentimentStrategy during a backtest.
-- Kept on the experiment row with the rest of its reproducibility snapshot.
ALTER TABLE experiments
    ADD COLUMN IF NOT EXISTS sentiment_models TEXT NOT NULL DEFAULT '[]';
