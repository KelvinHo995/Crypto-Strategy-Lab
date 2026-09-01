# Backend implementation review

Updated: 2026-09-01. Scope: fullstack runtime integration, market data,
experiment execution and documentation.

## Implemented and corrected

- Experiment correctness: validate backtest configuration, handle price gaps at
  candle open, prevent same-candle time-travel re-entry, reject invalid strategy
  configuration and keep deterministic ranking.
- Market data: Binance REST pagination and closed-candle normalization, Binance
  kline WebSocket reconnect/backoff, Postgres candle range/upsert repository and
  a two-year idempotent backfill command.
- Search/runtime: production search reads its requested candle range from
  Postgres, validates strict bounded JSON, uses three workers, persists state
  transitions and publishes progress/leaderboard events.
- Authentication: Postgres users, bcrypt password hashes, signed HS256 JWT with
  one-hour expiry, httpOnly SameSite=Lax cookie, protected routes and logout.
  The cookie receives the Secure flag when the request runs over TLS.
- Realtime: authenticated `/ws` emits subscribed `CANDLE_UPDATE`/`TRADE_TICK`
  plus global `SEARCH_PROGRESS`/`LEADERBOARD_UPDATE`. Trade traffic has a
  separate bounded queue so slow clients cannot crowd out higher-priority state.
- Sentiment: deterministic `crypto-lexicon/v1` model with validated FastAPI
  contract plus a Go client that reports service-down/invalid-response errors.
- Reliability: graceful HTTP shutdown, worker cancellation/wait, bounded request
  bodies, database query timeout in leaderboard notification and JSON-safe
  Postgres provenance binding.
- Final pre-commit review: normalize market symbols and strategy names at the
  HTTP boundary, normalize usernames case-insensitively, reject malformed live
  Binance numeric payloads, fail-fast on invalid live subscriptions and preserve
  the cookie `Secure` attribute during HTTPS logout.

## Project impact

- `JWT_SECRET` (minimum 16 characters) is now required to start the server; the
  agreed team value in `.env.example` is accepted.
- Production search requires candles to be backfilled; fewer than 21 candles in
  the requested range returns HTTP 422 instead of silently using fabricated data.
- Two shared Binance combined sockets cover eight symbols, four candle
  timeframes and aggregate trades. Browser clients subscribe over the one
  authenticated application WebSocket.
- Existing database tables are reused; candle writes remain idempotent under the
  existing `(symbol,timeframe,open_time)` key, so no migration was added.
- Two direct dependencies were added: `golang.org/x/crypto` for bcrypt and
  `github.com/coder/websocket` for WebSocket transport.

## Verification evidence

- `go test -count=3 ./...`: passed for auth, experiment, HTTP, market, sentiment
  client and all strategy packages.
- `go vet ./...`: passed.
- `go build ./cmd/server ./cmd/backfill`: passed.
- `python -m unittest discover -s tests -v`: passed with the bundled Python
  runtime (lexicon positive/negative/neutral behavior).
- `git diff --check`: passed; only Git's Windows LF-to-CRLF notices were emitted.
- Frontend `pnpm lint` and `pnpm build`: passed during the cross-project review;
  no UI/UX source was edited in the final review pass.

## Deliberate MVP boundaries

- One search request currently creates one candidate. Multi-candidate batches,
  genetic search and user cancellation remain stretch goals in ADR-0011.
- The sentiment service is lightweight and deterministic; FinBERT remains an
  explicitly documented later replacement behind the same contract.
- The catalog is intentionally fixed to eight liquid USDT pairs for this MVP;
  adding runtime exchange discovery or arbitrary user-entered pairs remains out
  of scope.
- End-to-end startup still requires external PostgreSQL/Binance availability and
  valid environment configuration; automated tests isolate those dependencies.
- The Go race detector was not available in the current Windows toolchain because
  CGO is disabled. Standard tests were repeated three times, but this is not a
  substitute for a future `go test -race ./...` run in a CGO-enabled environment.
- The agreed JWT value is intentionally retained for team compatibility. Because
  `.env.example` is public, it must be replaced with a private random value before
  any Internet-facing production deployment where forged sessions would matter.

## Fullstack integration update

- Added authenticated `GET /candles` for chart bootstrap data.
- Added frontend Auth Gate and same-origin Vite REST/WebSocket proxy.
- Connected strategy registry, historical charts, search, experiment leaderboard,
  realtime candles and search/leaderboard events to the Go backend.
- Aligned every frontend market selector with `GET /markets` and the backend
  timeframes (`5m`, `15m`, `1h`, `4h`), including charts, aggregate trades,
  strategy discovery and backtests.
- Removed silent market-data fallback in authenticated mode. Generated candles
  and trades now run only in explicit offline demo mode; the News collector and
  aggregate remain clearly labelled demo data.
- Updated the previously stale pnpm lockfile and removed frontend lint errors.

## 2026-09-01 teammate-integration follow-up

- Removed a committed `cookies.txt` session artifact and ignored future cookie
  exports; restored one canonical root `README.md` to avoid Windows casing
  collisions.
- Added observer-level panic isolation so a faulty WebSocket/observer callback
  cannot terminate a worker, including a regression test proving the worker
  completes the next queued job.
- Added compatibility decoding plus migration `0003` for two legacy experiment
  rows stored as pgx bytea-style hex text in Supabase. `GET /experiments` now
  decodes every existing row.
- Changed candle persistence from one SQL round-trip per row to idempotent
  500-row batches and made `cmd/backfill` load the documented local `.env`.
  A real Binance/Supabase run stored 210,239 `5m`, 70,079 `15m`, 17,519 `1h`,
  and 4,379 `4h` closed candles.
- Connected the News screen's sample-analysis action to the authenticated Go
  sentiment endpoint. Live observations are labelled `LIVE`; fixture feed and
  aggregate panels remain labelled `DEMO` because a production collector is
  outside MVP.
- Runtime smoke evidence: auth succeeded, six strategies listed, 167 recent
  1h candles loaded, search returned HTTP 202 then `COMPLETED`, experiments
  returned HTTP 200, WebSocket delivered `SEARCH_PROGRESS`, live sentiment
  returned HTTP 201 through the Vite proxy, and service-down returned HTTP 502.
- Final regression: `go test -count=3 ./...`, `go vet ./...`, frontend lint/build,
  Python sentiment tests, and `git diff --check` passed. The Go race detector
  remains unavailable because the current Windows toolchain has CGO disabled.

## 2026-09-01 multi-market runtime follow-up

- Added the authenticated `GET /markets` catalog for BTC, ETH, BNB, SOL, XRP,
  ADA, DOGE and AVAX against USDT. Unsupported candle and search symbols now
  fail at the HTTP boundary.
- Added multi-symbol backfill configuration, request pacing and strict invalid
  symbol handling. All eight markets were backfilled for 180 days across all
  four timeframes, then refreshed for the latest two days. The new ADA, DOGE and
  AVAX runs each stored 51,839 `5m`, 17,279 `15m`, 4,319 `1h`, and 1,079 `4h`
  candles before the latest refresh upsert.
- Fixed Binance kline decoding for the simultaneous lowercase `l` (low price)
  and uppercase `L` (last trade ID) fields. Without the explicit `L` field, Go's
  case-insensitive JSON matching could reject real Binance candles even though
  simplified fixtures passed.
- Added shared combined kline/aggregate-trade streams, per-client subscriptions,
  reconnect subscription restoration and a React Strict Mode cleanup fix for
  lightweight-charts and WebSocket socket replacement.
- Runtime evidence: authenticated candle queries returned 499 recent `1h`
  candles for every catalog symbol; an ADAUSDT MA+RSI search returned HTTP 202
  and reached `COMPLETED`; browser checks showed all eight choices in Strategy
  and Backtest selectors, four `LIVE/API` charts and a continuously updating
  real Binance trade feed.
- Final verification after all fixes: full `go test -count=3 ./...` and
  `go vet ./...` passed; frontend ESLint, TypeScript and Vite production build
  passed; Python sentiment unittest passed; `git diff --check` reported no
  whitespace errors. Browser verification also confirmed that the Backtest
  default range tracks the latest 90 days instead of a stale hard-coded year.

## 2026-09-01 caching and durable-queue follow-up

- Added bounded process-local caches at repository boundaries: candle range LRU
  with concurrent miss coalescing, write-invalidated leaderboard snapshot, and
  sentiment timestamp memoization. Errors are never cached and generation checks
  prevent an older in-flight read from repopulating stale data.
- Added migration `0005` for the leaderboard score index and bounded Postgres
  reads to the Top 100 before the application emits its Top 10 WebSocket view.
- Replaced the production in-memory channel with `PostgresQueue`. The initial
  experiment result and compact job commit atomically; workers use `SKIP LOCKED`,
  retry/backoff, Ack/Nack, renewable leases and stale claim recovery. Candle
  arrays remain in Postgres instead of being serialized into queue messages.
  Runtime testing exposed and fixed an integer-overflow in the millisecond
  backoff expression by making the arithmetic explicitly `BIGINT`; retry
  exhaustion now marks both the retained diagnostic job and experiment failed.
- Kept `InMemoryQueue` for unit tests and deliberately did not add Redis/Kafka;
  the existing Supabase database satisfies the current durability and
  multi-worker drivers without another operational dependency.
- Added `cmd/migrate` and applied idempotent migrations `0001` through `0007` to
  the configured Supabase project. The Postgres runtime integration test proved
  atomic PENDING+job persistence, compact claim payload decoding, Ack cleanup,
  and exhausted-retry propagation to the experiment status. Migration `0007`
  reconciles previously out-of-band `search_id`/`search_total` columns with the
  source model and gives fresh databases the same schema.
- Final live E2E after migration and server restart: register/login returned
  201/200, the authenticated catalog returned eight markets and six strategies,
  and every catalog symbol returned 491 recent 1h candles. Durable ADAUSDT
  MA+RSI search `exp-1788276798188744300` reached `COMPLETED` with 13 trades and
  persisted `searchId` plus `searchTotal: 1`; unsupported pair and missing-data
  requests returned HTTP 400 and 422. Sentiment analysis persisted successfully.
  Browser verification showed four `LIVE/API` charts, `WS Connected`, Binance
  API feed state, and continuously updating real aggregate trades.
