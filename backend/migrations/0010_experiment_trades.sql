CREATE TABLE IF NOT EXISTS experiment_trades (
    id                BIGSERIAL PRIMARY KEY,
    experiment_id     TEXT NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    pair              TEXT NOT NULL,
    entry_time        BIGINT NOT NULL,
    direction         TEXT NOT NULL,
    volume_usd        DOUBLE PRECISION NOT NULL,
    entry_price       DOUBLE PRECISION NOT NULL,
    stop_loss         DOUBLE PRECISION NOT NULL,
    take_profit       DOUBLE PRECISION NOT NULL,
    exit_price        DOUBLE PRECISION NOT NULL,
    exit_time         BIGINT NOT NULL,
    transaction_cost  DOUBLE PRECISION NOT NULL,
    slippage          DOUBLE PRECISION NOT NULL,
    profit            DOUBLE PRECISION NOT NULL
);

CREATE INDEX IF NOT EXISTS experiment_trades_experiment_id_idx ON experiment_trades (experiment_id);
