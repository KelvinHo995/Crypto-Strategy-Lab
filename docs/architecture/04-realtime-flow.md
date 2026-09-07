# 4. Realtime Data Flow

Live price updates must reach up to 4 independently-configured charts on the
frontend without the frontend polling and without coupling the frontend to
Binance's wire format (spec ch.4 "Yêu cầu kiến trúc").

## Flow

```
Binance combined WebSockets
     │  one kline connection + one aggregate-trade connection
     │
     ▼
internal/market — Binance Adapter
     │  StreamMarketEvents(catalog symbols, timeframes) → chan LiveEvent
     │  (candle.IsClosed = false while still forming, true once the
     │   interval closes — see contracts.md)
     ▼
cmd/server — /ws handler
     │  CANDLE_UPDATE and TRADE_TICK
     │  per-client subscription filtering + separate bounded trade queue
     ▼
WebSocket (Go ↔ React, single connection, multiple message types)
     │
     ▼
Frontend — shared WS client
     │  sends candle/trade subscribe commands and restores them after reconnect
     │  demuxes by type to chart(s) and recent-trade panel
     ▼
Chart component(s) — up to 4, independently timeframe-switchable
```

The frontend opens the socket only after login so the browser includes the
httpOnly JWT cookie. During Vite development, same-origin proxy routes forward
REST and `/ws` to the Go server. `GET /markets` supplies all selectors with the
same eight-symbol/four-timeframe catalog; unsupported pairs/timeframes are
rejected by the backend.

The same WebSocket connection also carries `SEARCH_PROGRESS` and
`LEADERBOARD_UPDATE` messages (see
[06-search-backtest-flow.md](06-search-backtest-flow.md)) — one connection,
tagged message types, not one socket per concern. A browser sends
`SUBSCRIBE_CANDLES` for each visible chart and `SUBSCRIBE_TRADES` for the chosen
trade feed. Switching a selector updates the subscription without reconnecting.
The backend still owns only two Binance sockets globally, regardless of the
number of authenticated browser clients.

## Why the frontend never touches Binance directly

Spec ch.4 explicitly calls this out as the required shape:

```
Frontend → Market Data Service → Binance Adapter → Binance   (required)
Frontend → Binance API                                        (rejected)
```

The adapter boundary means a second exchange (OKX, Bybit) can be added as a
second implementation behind the same internal interface without the
frontend or WebSocket message shape changing at all — see
[ADR-0001](../adr/0001-market-data-adapter.md) and spec ch.40 Q3 ("Market
Data Provider mới ... có phải sửa frontend không?" — answer must be no).

## Reliability

If either Binance WebSocket disconnects, `internal/market` reconnects with
bounded exponential backoff and resumes delivery without crashing the server.
Deterministic local-WebSocket tests cover the production
`StreamMarketEvents` disconnect/resumption path. One limitation remains: the
backend does not publish upstream Binance reconnect state as a browser event or
metric, so a browser still connected to the Go WebSocket cannot distinguish an
upstream pause from a quiet market.

Aggregate trades are substantially busier than candles, so each browser client
has a dedicated bounded trade queue. Dropping an old trade tick for a slow client
is acceptable; allowing that traffic to delay candle/search/leaderboard state is
not. The frontend retains the newest 50 ticks for the selected symbol.
