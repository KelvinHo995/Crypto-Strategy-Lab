-- sentiment_results was originally scoped as a pure observation projection
-- (sentiment output only, for SentimentStrategy's time-aligned lookup —
-- ADR-0006). The MVP spec's News requirement ("Collect -> Store -> Analyze
-- sentiment") needs "Store" to mean the collected article is actually
-- readable, not just an opaque news_id and a score. Adds the article
-- metadata the collector already has in hand at ingestion time.
ALTER TABLE sentiment_results ADD COLUMN IF NOT EXISTS title TEXT NOT NULL DEFAULT '';
ALTER TABLE sentiment_results ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT '';
ALTER TABLE sentiment_results ADD COLUMN IF NOT EXISTS url TEXT NOT NULL DEFAULT '';
