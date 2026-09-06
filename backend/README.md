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
`SENTIMENT_SERVICE_URL` defaults to `http://localhost:8000`.
`BACKTEST_WORKERS` controls worker-pool concurrency (default `3`, allowed
`1`–`64`). Apply migrations with `go run ./cmd/migrate` before starting. It runs all numbered migrations
in order. Migration `0003` is an
idempotent cleanup for experiment rows written by an older pgx binding;
`0004` adds the experiment update timestamp used by stale-job recovery,
`0005` indexes leaderboard scoring, `0006` creates the durable job queue,
`0007` normalizes per-search metadata, `0008` preserves legacy strategies while
moving them to per-instance JSON, `0009` stores article publication metadata,
`0010` stores experiment trades, `0011` stores normalized news, `0012` stores
sentiment model provenance, `0013` stores pair/timeframe, and `0014` backfills
only market coordinates that can be recovered without guessing.

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

Ingest recent RSS news, persist normalized articles, and enrich new articles
through the existing sentiment service with:

```bash
go run ./cmd/news-ingest
```

Set `NEWS_RSS_FEEDS` to a comma-separated list of RSS 2.0 URLs.
`NEWS_LOOKBACK` controls the fetch window and defaults to `24h`. The command is
safe to rerun: `news_items` is upserted independently, while stable article IDs
already in `sentiment_results` are skipped before analysis.

For a presentation without depending on the current news cycle, set
`NEWS_DEMO_RSS_FIXTURES=true`. The command starts a temporary valid RSS feed
containing positive, negative and neutral articles and sends it through the
same RSS parser, PostgreSQL repositories and sentiment HTTP service. Its source
is `Crypto Strategy Lab Demo RSS`, and the frontend labels those rows `DEMO`;
they are never presented as real reporting.

## Structure
- `cmd/server` — entrypoint, HTTP/WebSocket wiring only, no business logic
- `cmd/migrate` — idempotent ordered PostgreSQL/Supabase migration runner
- `internal/market` — Candle, exchange adapters (owner: Person 1)
- `internal/strategy` — Strategy interface, plugin registry, MA/RSI/Bollinger/SR/SMC/MACD/Sentiment and generators (owner: Person 2)
- `internal/experiment` — Backtester, Evaluator, weighted Ranking, durable Queue + configurable worker pool, runtime metrics and Result provenance (owner: Person 3)

`POST /search/start` is asynchronous: a successful request returns HTTP 202
with `STARTED`; use `GET /experiments/{searchId}` to observe
`PENDING|RUNNING|COMPLETED|FAILED`.

Authenticated frontend endpoints include `GET /markets`, `GET /strategies`,
`GET /experiments`, `GET /experiments/{id}/trades`, `POST /search/start`,
`GET /candles`, `POST /sentiment/analyze`, `GET /metrics`,
`GET /metrics/prometheus`, and `GET /ws`.
`GET /health` is process liveness; `GET /ready` checks PostgreSQL readiness.
Run the backfill before opening charts or starting searches so the requested
candle ranges exist.

The server opens two shared Binance combined streams: klines for every
supported symbol/timeframe and aggregate trades for every supported symbol.
Clients subscribe over the single authenticated `/ws` connection with
`SUBSCRIBE_CANDLES`/`UNSUBSCRIBE_CANDLES` and
`SUBSCRIBE_TRADES`/`UNSUBSCRIBE_TRADES`; this prevents every browser from
opening its own Binance connections.

Production search uses `PostgresQueue`: the PENDING result and compact job
payload are committed atomically, `BACKTEST_WORKERS` workers claim with `SKIP LOCKED`, and
leases are heartbeated/reclaimed after a crash. Candle payloads stay in
Postgres and are loaded through a bounded 60-second LRU cache. Leaderboard
snapshots use a write-invalidated three-second cache; sentiment timestamp
lookups use a bounded five-minute memoizer. These are process-local caches —
Postgres remains the source of truth and no Redis dependency is required.
`GET /metrics` exposes process-lifetime completed/failed counts, throughput,
average queue wait/execution time and live queue state. For the repeatable
in-memory architecture proof comparing worker counts, run:

```bash
go run ./cmd/perf -candidates 1000 -candles 2000 -workers 1,3 -repeat 3
```

This isolates worker-pool/backtest throughput from database and network
variance; it performs a discarded warm-up and reports per-run plus median
throughput/speedup. Production `/metrics` is the complementary end-to-end
measurement, while `/metrics/prometheus` exposes the same core signals in
Prometheus text format.
Queue integration tests operate on the configured database and are opt-in to
avoid competing with live workers: stop the backend and run
`$env:RUN_POSTGRES_QUEUE_INTEGRATION='1'; go test ./internal/experiment -run TestPostgresQueue`
in PowerShell.

The PostgreSQL trade and sentiment runtime tests are also opt-in. Run them
against a disposable database with `RUN_POSTGRES_TRADES_INTEGRATION=1` and
`RUN_POSTGRES_SENTIMENT_INTEGRATION=1`. Ordinary `go test ./...` never reaches
the configured Supabase database. Runtime tests delete their own rows before
closing the connection and report cleanup failures instead of leaving fixture
experiments in the leaderboard.

## Contracts

The shared contracts are defined in [`PLAN.md`](../PLAN.md#3-contract--interface-dùng-chung)
and elaborated in [`docs/architecture`](../docs/architecture/README.md). Go types
under `internal/market`, `internal/strategy`, and `internal/experiment` are the
executable source of truth and must remain compatible with those documents.
