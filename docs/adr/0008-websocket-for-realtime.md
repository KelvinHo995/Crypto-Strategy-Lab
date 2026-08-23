# ADR-0008: WebSocket for realtime UI, not polling/SSE

**Status:** Accepted
**Owner:** Market Data / Frontend

## Context

The frontend needs live-updating candlestick charts (up to 4 concurrent,
independently timeframe-switchable — spec ch.5), plus live search progress
and leaderboard updates as a search loop runs (spec ch.33 step 8-9). Spec
ch.4 explicitly shows repeated polling (`GET /price`, `GET /price`, `GET
/price`) as the pattern to avoid — "Frontend cần nhận cập nhật mà không liên
tục gọi." Some mechanism for the backend to push updates as they happen is
required.

## Decision

A single WebSocket connection per frontend session, carrying multiple
tagged message types over one channel:

```json
{ "type": "CANDLE_UPDATE",      "payload": { /* Candle */ } }
{ "type": "SEARCH_PROGRESS",    "payload": { "tested": 125, "total": 500 } }
{ "type": "LEADERBOARD_UPDATED","payload": { /* ExperimentResult[] */ } }
```
(full contract in `PLAN.md` §3.5)

The connection is bidirectional — the frontend can also send commands
(subscribe to a symbol+timeframe, `START_SEARCH`) over the same connection
that delivers updates back.

## Alternatives considered

- **Polling** (`GET /price` on an interval). Rejected — explicitly called
  out as the anti-pattern in spec ch.4 itself; wastes requests, latency is
  bounded by poll interval not by when data actually changes, and gets
  worse linearly with each additional concurrently-open chart.
- **Server-Sent Events (SSE).** A real contender — simpler than WebSocket
  (plain HTTP, no upgrade handshake, browser `EventSource` has built-in
  auto-reconnect). Rejected because SSE is one-directional (server→client
  only); the frontend also needs to *send* commands (subscribe to a new
  symbol/timeframe, `START_SEARCH`), which would need a second mechanism
  (plain POST requests) alongside SSE — two channels instead of one for a
  need WebSocket already covers in a single connection.
- **Long polling.** Rejected — more manual complexity to implement
  correctly (timeout handling, immediate re-issue on response) than
  WebSocket, for no real benefit over it at this message frequency.
- **One WebSocket connection per message type** (separate sockets for
  candles vs. leaderboard vs. search progress). Rejected — more connections
  to manage, more reconnect/backoff logic multiplied by N, for no benefit
  at this message volume; a single tagged-message connection is simpler and
  sufficient.

## Consequences

- **Positive:** one connection, one place to implement reconnect/backoff
  logic on the frontend — not N.
- **Positive:** bidirectional by default — command-and-update flows
  (`START_SEARCH` → `SEARCH_PROGRESS` → `LEADERBOARD_UPDATED`) stay on one
  channel.
- **Cost:** unlike SSE, WebSocket reconnect is not automatic — the frontend
  must hand-roll reconnect/backoff logic itself (this is separate from, but
  related to, the Binance-side reconnect logic in
  [04-realtime-flow.md](../architecture/04-realtime-flow.md), which is a
  different leg of the pipeline).
- **Cost, not a current driver:** load-balancing WebSocket connections
  across multiple backend instances is harder than stateless HTTP (sticky
  sessions or a shared pub/sub layer would be needed). Not relevant under
  the current single-instance modular monolith
  ([ADR-0005](0005-modular-monolith-not-microservices.md)) — worth
  revisiting only if that changes.
