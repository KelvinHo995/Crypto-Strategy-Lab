# backend

Go API server — Market Data, Strategy, Experiment domains.

## Run

Create the local environment file and replace its placeholder values:

```bash
cp .env.example .env
```

Set `DATABASE_URL` and a `JWT_SECRET` of at least 16 characters, then start the
server:

```bash
go run ./cmd/server
```

The server loads `.env` for local development without overriding variables
already set in the process environment. `SENTIMENT_SERVICE_URL` defaults to
`http://localhost:8000`. Apply the migrations before starting.

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

Authenticated frontend endpoints include `GET /strategies`, `GET /experiments`,
`POST /search/start`, `GET /candles`, and `GET /ws`. Run the backfill before
opening charts or starting searches so the requested candle ranges exist.

## Contracts

The shared contracts are defined in [`PLAN.md`](../PLAN.md#3-contract--interface-dùng-chung)
and elaborated in [`docs/architecture`](../docs/architecture/README.md). Go types
under `internal/market`, `internal/strategy`, and `internal/experiment` are the
executable source of truth and must remain compatible with those documents.
