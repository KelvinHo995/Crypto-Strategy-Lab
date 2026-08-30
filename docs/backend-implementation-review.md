# Backend implementation review

Updated: 2026-08-30. Scope: backend, sentiment service and documentation only;
no UI/UX source was changed in this review pass.

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
- Realtime: authenticated `/ws` emits `CANDLE_UPDATE`, `SEARCH_PROGRESS` and
  `LEADERBOARD_UPDATE`. Slow clients may lose intermediate events rather than
  block market ingestion; the next event carries current state.
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
- Four BTCUSDT streams (`5m`, `15m`, `1h`, `4h`) start with the server and share
  the authenticated WebSocket endpoint.
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
- Live server startup currently subscribes to BTCUSDT only. Multi-pair dynamic
  subscriptions are outside the current runtime contract.
- End-to-end startup still requires external PostgreSQL/Binance availability and
  valid environment configuration; automated tests isolate those dependencies.
- The Go race detector was not available in the current Windows toolchain because
  CGO is disabled. Standard tests were repeated three times, but this is not a
  substitute for a future `go test -race ./...` run in a CGO-enabled environment.
- The agreed JWT value is intentionally retained for team compatibility. Because
  `.env.example` is public, it must be replaced with a private random value before
  any Internet-facing production deployment where forged sessions would matter.
