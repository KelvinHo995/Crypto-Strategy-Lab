# 4. Realtime Data Flow

Live price updates must reach up to 4 independently-configured charts on the
frontend without the frontend polling and without coupling the frontend to
Binance's wire format (spec ch.4 "Yêu cầu kiến trúc").

## Flow

```
Binance WebSocket (kline stream)
     │
     ▼
internal/market — Binance Adapter
     │  StreamLiveCandles(symbol, timeframe) → chan Candle
     │  (candle.IsClosed = false while still forming, true once the
     │   interval closes — see contracts.md)
     ▼
cmd/server — /ws handler
     │  wraps each Candle in { "type": "CANDLE_UPDATE", "payload": Candle }
     ▼
WebSocket (Go ↔ React, single connection, multiple message types)
     │
     ▼
Frontend — shared WS client
     │  demuxes by "type", routes CANDLE_UPDATE to the chart(s)
     │  subscribed to that symbol+timeframe
     ▼
Chart component(s) — up to 4, independently timeframe-switchable
```

The frontend opens the socket only after login so the browser includes the
httpOnly JWT cookie. During Vite development, same-origin proxy routes forward
REST and `/ws` to the Go server. The four initial charts are aligned to
BTCUSDT `5m`, `15m`, `1h`, and `4h`; unsupported pairs/timeframes are not shown.

The same WebSocket connection also carries `SEARCH_PROGRESS` and
`LEADERBOARD_UPDATE` messages (see
[06-search-backtest-flow.md](06-search-backtest-flow.md)) — one connection,
tagged message types, not one socket per concern. The backend streams the four
MVP timeframes (`5m`, `15m`, `1h`, `4h`) over that connection; each chart
filters the timeframe it displays, so switching a chart does not reconnect.

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

If Binance's WebSocket disconnects, `internal/market` owns reconnect/retry —
this must not surface as a crash or a silent freeze on the frontend. This is
one of the explicit "vấn đáp" scenarios the team must be able to answer
(spec ch.32.4, ch.40 Q7, PLAN.md §5 checklist item 8). The chart should show
a visible "reconnecting" state rather than silently showing stale data as
live.
