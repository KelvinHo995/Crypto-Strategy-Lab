# 1. System Context

## Purpose

Crypto Strategy Lab is **not** a system for proving that a specific trading
strategy is profitable. It's a platform for testing, combining, and ranking
trading strategies systematically. The deliverable being graded is the
*software architecture* that lets strategies, search algorithms, and data
providers be added, replaced, or scaled without rewriting the rest of the
system — see PDF spec ch.2 and ch.47.

The system:
1. Pulls live and historical crypto price data from Binance.
2. Renders up to 4 independent-timeframe candlestick charts.
3. Runs pluggable technical-analysis strategies (MA, RSI, Bollinger,
   Support/Resistance, later MACD/SMC/Wyckoff/Sentiment) against that data.
4. Combines multiple strategies into composite strategies (majority vote or
   weighted score).
5. Backtests candidate strategies against historical data and scores them
   (Return, Win Rate, Max Drawdown, Trade Count).
6. Continuously searches the combination space (Random Search → Domain-guided
   → pluggable in future) and maintains a ranked Top-K leaderboard.
7. Collects crypto-related news and scores its sentiment (POSITIVE / NEGATIVE
   / NEUTRAL), which can itself feed a strategy.

## Primary actor

A user exploring the dashboard: picks a pair/timeframe, turns strategies on,
starts a search, watches the leaderboard update, inspects a winning
strategy's trades on the chart. This is also the exact flow the graded demo
follows (spec ch.46) and the flow `docs/architecture/06-search-backtest-flow.md`
describes end to end.

## External systems

| System | Direction | Notes |
|---|---|---|
| **Binance** (REST + WebSocket) | inbound | Only external market data source for MVP. Accessed only through `internal/market`'s adapter — nothing else in the codebase is allowed to import a Binance client directly (see [ADR-0001](../adr/0001-market-data-adapter.md)). |
| **News sources** (RSS / News API / crawler) | inbound | Abstracted behind a `NewsProvider` boundary so the source can change without touching sentiment analysis or strategies (spec ch.28). |

The sentiment model itself is **not** an external system — it's `sentiment-service`,
a component we own and deploy, reached over REST (see [ADR-0006](../adr/0006-separate-sentiment-service.md)).

## System boundary

```
                         ┌───────────────────────────┐
   Binance  ───REST/WS──▶│                           │
                         │      Crypto Strategy Lab    │◀──HTTP── User (browser)
   News source ───HTTP──▶│  (backend + frontend +      │
                         │   sentiment-service)         │
                         └───────────────────────────┘
```

## Explicit non-goals

- No real trading / order execution. Signals are advisory output only
  (spec ch.47: "Đồ án không nhằm chứng minh rằng MA+RSI+SMC có thể kiếm tiền thật").
- No multi-exchange support in MVP (Binance only); the architecture must
  *allow* adding OKX/Bybit/Coinbase adapters later without changing callers
  — it doesn't have to ship them.
- Minimal auth only — username/password + a short-lived (1h) signed JWT
  cookie, just enough to stop anonymous users from triggering backtests
  (real compute cost via the Job Queue). No roles, no password reset, no
  OAuth/SSO — see [ADR-0007](../adr/0007-simple-session-auth.md).
