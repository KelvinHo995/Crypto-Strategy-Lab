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
| `internal/market` | Market Data | `Candle` type, Binance adapter (REST historical + WS live), reconnect logic | Strategy logic, DB writes for anything but candles, chart-rendering concerns |
| `internal/strategy` | Strategy + Search | `Strategy` interface, `Signal` type, `Registry`, individual strategies (MA/RSI/BB/SR...), `StrategyGenerator` (Random/Domain-guided), `CandidateStrategy` composition | Binance calls, DB access, HTTP handlers — a strategy only sees candles in, signal out |
| `internal/experiment` | Experiment | Backtester (simulate trades), Evaluator (Return/WinRate/MDD/TradeCount), Ranking, `Queue` interface + `InMemoryQueue` + Worker pool for running many candidates concurrently ([ADR-0004](../adr/0004-inprocess-job-queue-not-kafka.md)), `Result` provenance ([ADR-0009](../adr/0009-experiment-provenance-storage.md)) | Strategy logic itself, market data fetching |
| `internal/auth` | shared infrastructure ([ADR-0007](../adr/0007-simple-session-auth.md)) | Implemented `User` storage, bcrypt hashing, JWT issuance/verification and Postgres repository; `internal/httpx` owns handlers/cookie middleware | Domain logic from any other package; this package knows nothing about strategies, candles, or experiments |
| `cmd/server` | shared (composition root) | HTTP/WebSocket route wiring only — constructs and injects the above, contains no business logic | Any domain logic; if `cmd/server` has an `if`/`switch` on strategy type, that's the God-Service anti-pattern (spec ch.44) |

This mirrors the codebase as it stands today: `market.Candle`,
`strategy.Strategy`/`strategy.Registry`, `experiment.Result` already exist as
separate packages with no import cycles back into each other except
`strategy` → `market` (for the `Candle` type) and `experiment` → `strategy`
(for `Signal`). `experiment` does **not** import `market` — it depends on
`strategy`'s interface, not on how candles are produced, which is the
"consumer defines the interface" Go idiom already used in
`experiment.Strategy`.

## Frontend feature packages

The current frontend MVP is a compact React dashboard in `frontend/src/App.tsx`
with six navigable product surfaces: Realtime, Strategy Engine, Discovery,
Backtest, News Crawler, and Settings. Shared visual rules live in `App.css`.
The screens currently use representative demo data while the remaining HTTP
and WebSocket contracts are implemented. When API integration begins, split
the surfaces into `features/{market,strategy,experiment,news}` and place the
API/WebSocket clients in `shared/`; that is the intended ownership boundary,
not a directory structure the repository already claims to have.

## Why this shape

Adding a strategy, a search algorithm, or a market data provider should each
touch exactly one package. If a change to "add MACD" requires editing
`cmd/server`, `internal/experiment`, and the frontend, that's a signal the
plugin boundary has leaked — this is the exact test the spec proposes in
ch.41 ("Scenario đánh giá khả năng mở rộng") and ch.40 Q1.
