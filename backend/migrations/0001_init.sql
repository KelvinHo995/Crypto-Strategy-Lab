-- Run once, manually, against the Supabase Postgres instance (ADR-0012).
-- Upsert-safe via primary keys — rerunning is fine.

CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS experiments (
    id                 TEXT PRIMARY KEY,
    candidate_id       TEXT NOT NULL,
    strategies         TEXT NOT NULL,
    params             TEXT NOT NULL,
    policy             TEXT NOT NULL,
    strategy_versions  TEXT NOT NULL,
    dataset_period     TEXT NOT NULL,
    return_pct         REAL,
    mdd                REAL,
    trade_count        INTEGER,
    win_rate           REAL,
    wins               INTEGER,
    losses             INTEGER,
    total_profit       REAL,
    status             TEXT NOT NULL,
    created_at         BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS candles (
    symbol     TEXT NOT NULL,
    timeframe  TEXT NOT NULL,
    open_time  BIGINT NOT NULL,
    open REAL, high REAL, low REAL, close REAL, volume REAL,
    PRIMARY KEY (symbol, timeframe, open_time)
);
