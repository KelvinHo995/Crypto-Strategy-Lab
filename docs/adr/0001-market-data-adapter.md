# ADR-0001: Market data behind a Provider/Adapter boundary

**Status:** Accepted
**Owner:** Market Data (Person 1)

## Context

Binance is the only market data source for MVP, but the spec explicitly
requires the architecture to tolerate a second exchange later without
rewriting consumers (PDF spec ch.4: "Nhờ đó sau này có thể bổ sung
BinanceAdapter, OKXAdapter, BybitAdapter, CoinbaseAdapter mà frontend không
phải thay đổi"). Every downstream consumer — strategies, the backtester, the
realtime chart, the WebSocket layer — needs candle data, but none of them
should need to know it came from Binance specifically.

## Decision

All Binance access (REST klines for history, WebSocket for live ticks) is
isolated inside `internal/market`, exposed only through the normalized
`Candle` type and adapter functions (`FetchHistoricalCandles`,
`StreamLiveCandles`). No other package — not `internal/strategy`, not
`internal/experiment`, not `cmd/server`, and definitely not the frontend —
imports a Binance client or knows Binance's wire format.

```
Frontend → Market Data Service → Binance Adapter → Binance     (required shape)
Frontend → Binance API                                          (rejected)
```

## Alternatives considered

- **Call Binance directly from wherever candles are needed** (e.g. the
  backtester fetches its own history, the WS handler streams straight from
  Binance to the frontend). Rejected: couples every consumer to Binance's
  response shape and rate limits; a second exchange would require changing
  every call site instead of one adapter.
- **Generic "any exchange" abstraction from day one** (plugin-loaded
  exchange configs, capability negotiation). Rejected as premature — MVP
  only needs one exchange; the adapter interface is deliberately just wide
  enough for a second implementation to slot in later, not a general
  exchange-abstraction framework nobody asked for yet.

## Consequences

- **Positive:** adding `OKXAdapter` later is additive — implement the same
  adapter surface, register it, nothing downstream changes. This is the
  concrete answer to spec ch.40 Q3.
- **Positive:** Binance-specific concerns (rate limiting, reconnect/backoff
  on WS drop) are contained in one place, which is also where reliability
  testing (spec ch.32.4) is scoped.
- **Cost:** one extra layer of indirection even though there's only one
  exchange today — accepted because the spec grades specifically on this
  boundary existing, not on multi-exchange support actually shipping.
