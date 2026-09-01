CREATE TABLE IF NOT EXISTS experiment_jobs (
    id           TEXT PRIMARY KEY,
    payload      JSONB NOT NULL,
    status       TEXT NOT NULL CHECK (status IN ('QUEUED', 'RUNNING', 'FAILED')),
    attempts     INTEGER NOT NULL DEFAULT 0,
    available_at BIGINT NOT NULL,
    locked_at    BIGINT,
    last_error   TEXT,
    created_at   BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS experiment_jobs_claim_idx
ON experiment_jobs (status, available_at, created_at)
WHERE status IN ('QUEUED', 'RUNNING');
