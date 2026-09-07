# 2. Module Decomposition

## Deployable units (C4 Container level)

Three separately runnable processes, matching the repo's top-level folders:

| Container | Tech | Responsibility |
|---|---|---|
| `backend/` | Go 1.26.5, `net/http` (no framework — [ADR-0005](../adr/0005-modular-monolith-not-microservices.md)) | Market data ingestion, news/sentiment orchestration, strategy engine, search, backtesting, experiment storage, leaderboard, WebSocket fan-out. A **modular monolith** whose domain packages are separated by package boundaries, not process boundaries. |
| `frontend/` | React + TypeScript (Vite, npm) | Dashboard: multi-timeframe charts, strategy picker, leaderboard, trade detail, news/sentiment panel. Talks to backend only via REST + one WebSocket — never talks to Binance or the sentiment service directly (spec ch.4 "Yêu cầu kiến trúc"). |
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
| `internal/strategy` | Strategy + Search | `Strategy` interface, `Signal` type, plugin descriptors/registry, seven concrete strategies, `RandomGenerator`, `StrategyGenerator` boundary, `CandidateStrategy` composition | Binance calls, DB access, HTTP handlers — a strategy only sees candles in, signal out |
| `internal/experiment` | Experiment | Backtester, Evaluator, cached/indexed leaderboard, `Queue` interface + durable `PostgresQueue` + panic-isolated Worker pool ([ADR-0013](../adr/0013-postgres-durable-queue-and-local-caches.md)), `Result` provenance ([ADR-0009](../adr/0009-experiment-provenance-storage.md)) | Strategy logic itself, exchange fetching |
| `internal/news` | News | Source-neutral `NewsItem`/`NewsProvider`, RSS normalization, stable IDs, related-coin extraction, normalized article persistence | Model inference, strategy decisions, frontend presentation |
| `internal/sentiment` | Market Data / Sentiment | Go client for FastAPI, ingestion orchestration, observation persistence and max-age-constrained time lookup used by `SentimentStrategy` | Python model implementation, frontend presentation, raw article content in `sentiment_results` |
| `internal/auth` | shared infrastructure ([ADR-0007](../adr/0007-simple-session-auth.md)) | Implemented `User` storage, bcrypt hashing, JWT issuance/verification and Postgres repository; `internal/httpx` owns handlers/cookie middleware | Domain logic from any other package; this package knows nothing about strategies, candles, or experiments |
| `cmd/server` | shared (composition root) | HTTP/WebSocket wiring and source-level strategy plugin registration — constructs and injects dependencies, contains no business logic | Strategy-specific generation/execution logic or other domain behavior |

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
data is confined to explicit offline `DEMO` mode. In live mode the News page
reads the last 24 hours of analyzed observations from
`GET /sentiment/observations`; it does not substitute demo data on an empty or
failed response. There is still no `GET /news`, so stored article content and
`relatedCoins` are not exposed to the frontend.

## Why this shape

Adding a strategy now touches its strategy-owned implementation/factory/random
parameter descriptor plus one source-level registration in `cmd/server`.
Generic generation, experiment, backtest, evaluation, and ranking code remains
unchanged. Search algorithms and market providers likewise change their
implementation and composition wiring, not their consumers. This is the exact
test the spec proposes in ch.41 ("Scenario đánh giá khả năng mở rộng") and
ch.40 Q1; runtime binary plugin discovery is not part of the design.
