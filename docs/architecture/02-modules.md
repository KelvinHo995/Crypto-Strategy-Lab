# 2. Module Decomposition

## Deployable units (C4 Container level)

Three separately runnable processes, matching the repo's top-level folders:

| Container | Tech | Responsibility |
|---|---|---|
| `backend/` | Go 1.22, `net/http` (no framework — [ADR-0005](../adr/0005-modular-monolith-not-microservices.md)) | Market data ingestion, strategy engine, search, backtesting, experiment storage, leaderboard, WebSocket fan-out. A **modular monolith**: one binary, three internal domains kept apart by package boundaries, not process boundaries. |
| `frontend/` | React + TypeScript (Vite, pnpm) | Dashboard: multi-timeframe charts, strategy picker, leaderboard, trade detail, news/sentiment panel. Talks to backend only via REST + one WebSocket — never talks to Binance or the sentiment service directly (spec ch.4 "Yêu cầu kiến trúc"). |
| `sentiment-service/` | Python 3.11, FastAPI, uv | Text → sentiment classification. The one component deliberately split out of the Go monolith, because it needs the Python ML ecosystem ([ADR-0006](../adr/0006-separate-sentiment-service.md)). |

```
Frontend (React)
     │ HTTP + WebSocket
     ▼
Backend (Go monolith)
     │ REST (POST /analyze)
     ▼
sentiment-service (FastAPI)

Backend
     │ REST + WS
     ▼
Binance
```

## Backend internal packages (C4 Component level)

`backend/internal/` — each package is owned by one person (see root README
"Team & Ownership"); cross-package changes need that owner's review.

| Package | Owner (domain) | Responsibility | Must NOT contain |
|---|---|---|---|
| `internal/market` | Market Data | Market catalog, `Candle`/`TradeTick` contracts, Binance adapter (REST historical + shared combined WS streams), reconnect logic | Strategy logic, DB writes for anything but candles, chart-rendering concerns |
| `internal/strategy` | Strategy + Search | `Strategy` interface, `Signal` type, `Registry`, individual strategies (MA/RSI/BB/SR...), `StrategyGenerator` (Random/Domain-guided), `CandidateStrategy` composition | Binance calls, DB access, HTTP handlers — a strategy only sees candles in, signal out |
| `internal/experiment` | Experiment | Backtester, Evaluator, cached/indexed leaderboard, `Queue` interface + durable `PostgresQueue` + panic-isolated Worker pool ([ADR-0013](../adr/0013-postgres-durable-queue-and-local-caches.md)), `Result` provenance ([ADR-0009](../adr/0009-experiment-provenance-storage.md)) | Strategy logic itself, exchange fetching |
| `internal/sentiment` | Market Data / Sentiment | Go client for FastAPI, observation persistence and bounded time lookup used by `SentimentStrategy` | Python model implementation, frontend presentation, raw article storage |
| `internal/auth` | shared infrastructure ([ADR-0007](../adr/0007-simple-session-auth.md)) | Implemented `User` storage, bcrypt hashing, JWT issuance/verification and Postgres repository; `internal/httpx` owns handlers/cookie middleware | Domain logic from any other package; this package knows nothing about strategies, candles, or experiments |
| `cmd/server` | shared (composition root) | HTTP/WebSocket route wiring only — constructs and injects the above, contains no business logic | Any domain logic; if `cmd/server` has an `if`/`switch` on strategy type, that's the God-Service anti-pattern (spec ch.44) |

This mirrors the codebase as it stands today: `market.Candle`,
`strategy.Strategy`/`strategy.Registry`, `experiment.Result` already exist as
separate packages with no import cycles back into each other except
`strategy` → `market` (for the `Candle` type) and `experiment` → `strategy`
(for `Signal` and composition). `experiment` also imports `market.Candle` as
the shared historical-data DTO used by backtests and queued jobs; it never
imports or calls the Binance adapter. Market acquisition therefore remains
behind `market.LiveProvider`/`market.CandleRepository` at the HTTP composition
boundary.

## Frontend feature packages

The React dashboard is split into `features/{market,strategy,experiment,news}`
with REST/WebSocket clients in `shared/`. Authentication, market catalog,
historical candles, aggregate trades, strategy registry, search/experiments,
sentiment analysis, and realtime events use backend contracts. Generated market
data is confined to explicit offline `DEMO` mode. The news collector/extraction
visuals and 24-hour sentiment aggregate remain labelled `DEMO`; they are not
presented as production data.

## Why this shape

Adding a strategy, a search algorithm, or a market data provider should each
touch exactly one package. If a change to "add MACD" requires editing
`cmd/server`, `internal/experiment`, and the frontend, that's a signal the
plugin boundary has leaked — this is the exact test the spec proposes in
ch.41 ("Scenario đánh giá khả năng mở rộng") and ch.40 Q1.
