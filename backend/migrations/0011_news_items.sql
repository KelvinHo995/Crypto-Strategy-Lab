CREATE TABLE IF NOT EXISTS news_items (
    id            TEXT PRIMARY KEY,
    title         TEXT NOT NULL,
    content       TEXT NOT NULL,
    source        TEXT NOT NULL,
    url           TEXT NOT NULL DEFAULT '',
    published_at  BIGINT NOT NULL,
    related_coins JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at    BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_news_items_published_at
    ON news_items (published_at DESC);
