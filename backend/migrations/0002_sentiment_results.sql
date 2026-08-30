CREATE TABLE IF NOT EXISTS sentiment_results (
    news_id       TEXT PRIMARY KEY,
    published_at  BIGINT NOT NULL,
    sentiment     TEXT NOT NULL CHECK (sentiment IN ('POSITIVE', 'NEGATIVE', 'NEUTRAL')),
    score         REAL NOT NULL CHECK (score >= 0 AND score <= 1),
    model_name    TEXT NOT NULL,
    model_version TEXT NOT NULL,
    analyzed_at   BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS sentiment_results_published_at_idx
    ON sentiment_results (published_at DESC);
