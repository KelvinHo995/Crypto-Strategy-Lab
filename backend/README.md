# backend

Go API server — Market Data, Strategy, Experiment domains.

## Run
```bash
go run ./cmd/server
```

Required environment: `DATABASE_URL` and a random `JWT_SECRET` of at least 32
characters. Apply `migrations/0001_init.sql` once before starting.

Backfill the fixed two-year dataset manually (safe to rerun):

```bash
go run ./cmd/backfill
```

## Structure
- `cmd/server` — entrypoint, HTTP/WebSocket wiring only, no business logic
- `internal/market` — Candle, exchange adapters (owner: Person 1)
- `internal/strategy` — Strategy interface, registry, generators (owner: Person 2)
- `internal/experiment` — Backtester, Evaluator, weighted Ranking, Queue + 3-worker pool, Result provenance (owner: Person 3)

`POST /search/start` is asynchronous: a successful request returns HTTP 202
with `STARTED`; use `GET /experiments/{searchId}` to observe
`PENDING|RUNNING|COMPLETED|FAILED`.

## Contracts

The shared contracts are defined in [`PLAN.md`](../PLAN.md#3-contract--interface-dùng-chung)
and elaborated in [`docs/architecture`](../docs/architecture/README.md). Go types
under `internal/market`, `internal/strategy`, and `internal/experiment` are the
executable source of truth and must remain compatible with those documents.
