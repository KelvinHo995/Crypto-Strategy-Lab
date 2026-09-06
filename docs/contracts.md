# Runtime contracts

This file is the concise runtime companion to `PLAN.md` §3. Go structs remain
the executable source of truth.

## Candle

`{symbol,timeframe,openTime,open,high,low,close,volume,isClosed}`. Historical
backfill persists only closed candles. Live updates may carry
`isClosed:false` until Binance closes the interval.

## Search and experiments

`GET /strategies` returns the authoritative names registered by the backend.
The frontend must render unknown names with a generic default-parameter form;
hardcoded metadata may enrich known plugins but must not hide a newly registered
one.

`POST /search/start` accepts `{pair,timeframe,from,to,capital,instances,policy}`,
where `instances` is `[{type,params,weight}]` — each entry is an
independently configured strategy (own params, own weight), so a composite
can combine more than one instance of the same type (e.g. MA(20) and MA(50))
without them sharing params. `policy` is `"majority"` (default) or
`"weighted"`; `weight` on each instance is only used under `"weighted"` and
is normalized server-side to sum to 1. Returns HTTP 202
`{searchId,status:"STARTED"}`. The result transitions through
`PENDING|RUNNING|COMPLETED|FAILED`. `GET /experiments` returns ranked results;
`GET /experiments/{id}` returns one provenance snapshot, including the exact
`instances` and `policy` that ran — not a lossy flat `strategies[]`+`params{}`
snapshot. Every persisted result also carries `searchId` and `searchTotal`;
the one-candidate flow uses the result ID and `1`.

Each new result also exposes `pair`, `timeframe`, and `datasetPeriod` (the
requested Unix-millisecond `from-to` range). Together these identify the exact
market candle dataset used by the backtest. Rows created before this provenance
was introduced omit `pair`/`timeframe` rather than claiming an unverified value.
Completed trade details remain available at `GET /experiments/{id}/trades`.
The endpoint returns HTTP 404 for an unknown experiment and an array (possibly
empty) for an existing experiment. A worker persists trades before publishing
`COMPLETED`; if trade persistence fails, that candidate is persisted and
broadcast as `FAILED`, so a reproducible result never advertises success
without its trade detail.

Completed results that consumed `SentimentStrategy` observations also expose
`sentimentModels:[{name,version}]`. The list contains every distinct persisted
model identity actually used during that backtest, sorted deterministically;
it is empty/omitted for non-sentiment runs or when sentiment fell back without
consuming a model output. `strategyVersions.Sentiment` remains the version of
the Go strategy implementation, not the Python model.

`POST /search/loop` accepts:

```json
{
  "pair": "BTCUSDT",
  "timeframe": "1h",
  "from": 1780000000000,
  "to": 1790000000000,
  "capital": 10000,
  "maxCandidates": 5,
  "maxDurationSeconds": 300,
  "noImprovementLimit": 3
}
```

`maxCandidates` is required and bounded to 2–200. `maxDurationSeconds` is `0`
or 60–3600; `noImprovementLimit` is non-negative. HTTP 202 returns
`{searchId,status:"STARTED",maxCandidates}`. HTTP 422 means the requested
pair/timeframe/range has fewer than `experiment.MinCandlesForBacktest`
(currently 202 — one more than the worst-case strategy lookback the
`RandomGenerator` can produce, `strategy.MaxGeneratedLookback`, itself tied to
`maLongWindow`'s 50–200 range) persisted candles, and must be backfilled;
the server never substitutes generated candles.

The stop rule is OR semantics: stop generating when the first enabled limit
is reached, while already-enqueued jobs finish normally. An early stop
(duration/no-improvement/cancellation/enqueue failure) is distinguishable
from a running search through WebSocket: the backend broadcasts a terminal
`SEARCH_PROGRESS{searchId,status:"STOPPED"|"FAILED",reason}` the instant
generation stops, so `{tested,total}` no longer has to reach the requested
maximum for a client to know the run is done generating. `status` is
`"FAILED"` only if nothing was ever enqueued. Still open: the production
queue has no backpressure, so a `noImprovementLimit` can overshoot by more
than a couple of candidates before the stop lands, and candidate dedup gives
up after 10 retries and may still enqueue a literal duplicate.

Production commits the initial `PENDING` row and compact `experiment_jobs`
message atomically for both `/search/start` and `/search/loop`, via
`EnqueuePending` — no crash window between the two writes. A job contains
dataset coordinates rather than candle rows; workers claim with a renewable
lease, load candles from the repository, and Ack/Nack after persistence.
`GET /experiments` returns at most the indexed, score-ranked Top 100;
WebSocket leaderboard updates publish Top 10.

## Historical market data

`GET /markets` returns the supported catalog. The current symbols are
`BTCUSDT`, `ETHUSDT`, `BNBUSDT`, `SOLUSDT`, `XRPUSDT`, `ADAUSDT`, `DOGEUSDT`,
and `AVAXUSDT`, each with `5m`, `15m`, `1h`, and `4h` timeframes.

`GET /candles?symbol=BTCUSDT&timeframe=5m&from=<unix-ms>&to=<unix-ms>&limit=500`
returns closed candles from Postgres. Supported MVP timeframes are `5m`, `15m`,
`1h`, and `4h`; `limit` is bounded to 1-5000. Unsupported symbols/timeframes
are rejected. Production search also rejects pairs outside this catalog and
returns HTTP 422 when the requested range contains fewer than
`experiment.MinCandlesForBacktest` candles (currently 202).

## Authentication

`POST /auth/register` creates a bcrypt-backed user. `POST /auth/login` sets an
httpOnly SameSite=Lax `session` JWT cookie with a one-hour expiry. The configured
secret must contain at least 16 characters. All routes
except health/readiness/register/login require that cookie. Logout clears it.

## Operational endpoints

`GET /health` is a dependency-free liveness probe. `GET /ready` calls
PostgreSQL with a two-second deadline and returns HTTP 503 when it is not
available. Authenticated `GET /metrics` returns:

```json
{
  "uptimeSeconds": 60,
  "workerCount": 3,
  "jobsCompleted": 20,
  "jobsFailed": 1,
  "jobsPerMinute": 21,
  "averageQueueWaitMs": 4.2,
  "queueWaitP50Ms": 3.8,
  "queueWaitP95Ms": 8.1,
  "averageExecutionTimeMs": 18.7,
  "executionTimeP50Ms": 17.9,
  "executionTimeP95Ms": 31.4,
  "queue": {"queued": 2, "running": 3, "failed": 1}
}
```

Counters and averages are process-lifetime measurements; queue state is queried
live from the active queue. Every HTTP response includes `X-Request-ID` and the
backend logs that ID with method, path, HTTP status, response bytes and elapsed
milliseconds.

Authenticated `GET /metrics/prometheus` exposes the same core worker, queue,
failure, throughput and p50/p95 latency signals in Prometheus text exposition
format. Normal asynchronous job transitions are correlated separately through
structured key-value log fields `search_id`, `job_id`, `candidate_id` and
`status`.

## WebSocket

`GET /ws` sends tagged JSON messages on one authenticated connection:

```json
{"type":"CANDLE_UPDATE","payload":{}}
{"type":"TRADE_TICK","payload":{"symbol":"BTCUSDT","tradeId":1,"tradeTime":0,"price":0,"quantity":0,"side":"BUY"}}
{"type":"SEARCH_PROGRESS","payload":{"tested":1,"total":5}}
{"type":"LEADERBOARD_UPDATE","payload":[]}
```

Search is started through REST; WebSocket is the server-push channel. Ordinary
per-candidate progress payload is `{tested,total,searchId}`; the terminal
early-stop broadcast additionally carries `{status,reason}` (`STOPPED` or
`FAILED`). The frontend renders `COMPLETED|STOPPED|FAILED` from these fields;
clients must not invent them or advance progress with timers. Reaching
`tested == total` prompts the client to fetch the persisted candidate and use
its real terminal status; it is not itself proof of `COMPLETED`. A client
selects only the market events it needs:

```json
{"type":"SUBSCRIBE_CANDLES","payload":{"symbol":"ETHUSDT","timeframe":"1h"}}
{"type":"UNSUBSCRIBE_CANDLES","payload":{"symbol":"ETHUSDT","timeframe":"1h"}}
{"type":"SUBSCRIBE_TRADES","payload":{"symbol":"SOLUSDT"}}
{"type":"UNSUBSCRIBE_TRADES","payload":{"symbol":"SOLUSDT"}}
```

Search and leaderboard events remain global to the authenticated connection.
Trade ticks use a separate bounded client queue so a busy aggregate-trade feed
cannot crowd out candles or experiment state updates.

The browser authenticates first, then opens `/ws` with its httpOnly session
cookie. Vite proxies REST and WebSocket paths to port 8080 in development.

## Sentiment

`POST /analyze` accepts `{newsId,text}` and returns
`{newsId,sentiment,score,model:{name,version},createdAt}`. The Go client treats
service errors explicitly; `SentimentStrategy` degrades to its base strategy.

The authenticated Go endpoint `POST /sentiment/analyze` accepts
`{newsId,text,publishedAt}` where `publishedAt` is Unix milliseconds. It sends
the text to the Python service and persists only the returned observation, not
the article text. HTTP 201 returns
`{newsId,publishedAt,sentiment,score,modelName,modelVersion,analyzedAt}`; an
unavailable/model-failing Python service returns HTTP 502. Strategy lookup uses the latest observation published at or
before the candle open time, bounded to the previous 24 hours to prevent
look-ahead and stale sentiment. The stored model confidence is translated to
the strategy's directional `0..1` contract: positive keeps the confidence,
negative uses `1-confidence`, and neutral is `0.5`.
