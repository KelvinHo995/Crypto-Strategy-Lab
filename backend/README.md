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

The server and backfill command load `.env` for local development without
overriding variables already set in the process environment.
`SENTIMENT_SERVICE_URL` defaults to `http://localhost:8000`. Apply migrations
with `go run ./cmd/migrate` before starting. It runs `0001` through `0007` in
order. Migration `0003` is an
idempotent cleanup for experiment rows written by an older pgx binding;
`0004` adds the experiment update timestamp used by stale-job recovery,
`0005` indexes leaderboard scoring, `0006` creates the durable job queue, and
`0007` normalizes per-search metadata (`search_id`, `search_total`).

Backfill the fixed two-year dataset manually (safe to rerun):

```bash
go run ./cmd/backfill
```

Backfill fetches all four MVP timeframes and upserts candles in batches of 500
to avoid one database round-trip per candle through the Supabase pooler.
Use `BACKFILL_SYMBOLS` for a comma-separated market list, `BACKFILL_DAYS` to
override the two-year default, and `BACKFILL_REQUEST_PAUSE_MS` to tune Binance
request pacing. `BACKFILL_SYMBOL` remains supported for older scripts. Invalid
symbols fail fast instead of silently filling a different market.

## Structure
- `cmd/server` — entrypoint, HTTP/WebSocket wiring only, no business logic
- `cmd/migrate` — idempotent ordered PostgreSQL/Supabase migration runner
- `internal/market` — Candle, exchange adapters (owner: Person 1)
- `internal/strategy` — Strategy interface, registry, generators (owner: Person 2)
- `internal/experiment` — Backtester, Evaluator, weighted Ranking, Queue + 3-worker pool, Result provenance (owner: Person 3)

`POST /search/start` is asynchronous: a successful request returns HTTP 202
with `STARTED`; use `GET /experiments/{searchId}` to observe
`PENDING|RUNNING|COMPLETED|FAILED`.

Authenticated frontend endpoints include `GET /markets`, `GET /strategies`,
`GET /experiments`, `POST /search/start`, `GET /candles`,
`POST /sentiment/analyze`, and `GET /ws`.
Run the backfill before opening charts or starting searches so the requested
candle ranges exist.

The server opens two shared Binance combined streams: klines for every
supported symbol/timeframe and aggregate trades for every supported symbol.
Clients subscribe over the single authenticated `/ws` connection with
`SUBSCRIBE_CANDLES`/`UNSUBSCRIBE_CANDLES` and
`SUBSCRIBE_TRADES`/`UNSUBSCRIBE_TRADES`; this prevents every browser from
opening its own Binance connections.

Production search uses `PostgresQueue`: the PENDING result and compact job
payload are committed atomically, three workers claim with `SKIP LOCKED`, and
leases are heartbeated/reclaimed after a crash. Candle payloads stay in
Postgres and are loaded through a bounded 60-second LRU cache. Leaderboard
snapshots use a write-invalidated three-second cache; sentiment timestamp
lookups use a bounded five-minute memoizer. These are process-local caches —
Postgres remains the source of truth and no Redis dependency is required.
Queue integration tests operate on the configured database and are opt-in to
avoid competing with live workers: stop the backend and run
`$env:RUN_POSTGRES_QUEUE_INTEGRATION='1'; go test ./internal/experiment -run TestPostgresQueue`
in PowerShell.

## Contracts

The shared contracts are defined in [`PLAN.md`](../PLAN.md#3-contract--interface-dùng-chung)
and elaborated in [`docs/architecture`](../docs/architecture/README.md). Go types
under `internal/market`, `internal/strategy`, and `internal/experiment` are the
executable source of truth and must remain compatible with those documents.
