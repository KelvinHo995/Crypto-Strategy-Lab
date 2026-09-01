# 3. Historical Data Flow (Backfill)

Historical candles are what every backtest runs against, and what indicators
are computed on. This flow is deliberately **not** on the request path — it's
a one-time (or periodic, manual) batch job, not something that runs on every
server start (see [PLAN.md](../../PLAN.md) §3c: "Backfill dataset: cố định 2
năm, không cho user chọn range tùy ý").

## Flow

```
Binance REST (klines)
     │  FetchHistoricalCandles(symbol, timeframe, range)
     ▼
internal/market — Binance Adapter
     │  → []Candle (normalized, exchange-agnostic shape)
     ▼
cmd/backfill (one-off CLI, one or many catalog symbols, run manually)
     │  batch upsert (500 rows) by primary key
     ▼
Supabase (Postgres) — candles table
     │  read-only from here on
     ▼
internal/experiment — Backtester
     (reads a candle range for a given symbol+timeframe+period)
```

## Why upsert-by-key, not append-only

`candles` is keyed on `(symbol, timeframe, open_time)`. Re-running
`cmd/backfill` is always safe — it overwrites the same rows rather than
duplicating them. This avoids needing gap-detection or dynamic-range logic:
the team explicitly decided that complexity isn't justified for a fixed
2-year, manually-triggered backfill (PLAN.md §3c). If the scope grows to
support arbitrary user-chosen date ranges, gap-checking becomes a real
requirement and this decision should be revisited via a new ADR — it is not
a design gap today, it's a documented scope cut.

`BACKFILL_SYMBOLS` selects a comma-separated subset of the catalog;
`BACKFILL_DAYS` optionally narrows the default two-year range. The command paces
Binance REST calls and rejects unsupported symbols before any writes. The
repository writes at most 500 candles per SQL statement. This keeps the
same transaction/idempotency semantics while avoiding hundreds of thousands
of network round-trips through the Supabase transaction pooler.

## What consumes this table

The search HTTP orchestrator reads the requested candle range through
`market.CandleRepository` and passes the normalized slice into
`internal/experiment`'s Backtester. The authenticated `GET /candles` endpoint
also reads this repository to seed frontend charts. In authenticated `LIVE`
mode, an unavailable API or missing range is a visible error; generated fixtures
are used only after the user explicitly enters offline `DEMO` mode. Live chart
updates do **not** go through this table — see
[04-realtime-flow.md](04-realtime-flow.md) for the separate, unbuffered path
live ticks take to the frontend. Indicator calculation (MA, RSI, Bollinger)
happens in `internal/strategy` over whatever candle slice it's given,
whether that slice came from the historical table (backtest) or the live
WebSocket buffer (realtime chart) — the strategy code itself doesn't know or
care which.

The server decorates Postgres with a 128-entry, 60-second process-local LRU.
Exact concurrent range requests share one database load. Returned slices are
cloned, writes advance an invalidation generation, and an older in-flight read
cannot reinsert stale data. The HTTP response retains a private 30-second browser
cache. Redis is unnecessary while there is one backend instance; Postgres
remains authoritative.

## Boundary enforced

Nothing outside `internal/market` constructs a Binance API request. Every
other package — including `cmd/backfill` — depends only on `market.Candle`
and the adapter's exported functions. See
[ADR-0001](../adr/0001-market-data-adapter.md) for why, and
`docs/architecture/02-modules.md` for the package ownership this protects.
