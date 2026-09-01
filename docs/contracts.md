# Runtime contracts

This file is the concise runtime companion to `PLAN.md` §3. Go structs remain
the executable source of truth.

## Candle

`{symbol,timeframe,openTime,open,high,low,close,volume,isClosed}`. Historical
backfill persists only closed candles. Live updates may carry
`isClosed:false` until Binance closes the interval.

## Search and experiments

`POST /search/start` accepts `{pair,timeframe,from,to,capital,strategies}` and
returns HTTP 202 `{searchId,status:"STARTED"}`. The result transitions through
`PENDING|RUNNING|COMPLETED|FAILED`. `GET /experiments` returns ranked results;
`GET /experiments/{id}` returns one provenance snapshot. Every persisted result
also carries `searchId` and `searchTotal`; the current one-candidate flow uses
the result ID and `1`, leaving an explicit compatible contract for future batch
discovery.

Production commits the initial `PENDING` row and compact `experiment_jobs`
message atomically. A job contains dataset coordinates rather than candle rows;
workers claim with a renewable lease, load candles from the repository, and
Ack/Nack after persistence. `GET /experiments` returns at most the indexed,
score-ranked Top 100; WebSocket leaderboard updates publish Top 10.

## Historical market data

`GET /markets` returns the supported catalog. The current symbols are
`BTCUSDT`, `ETHUSDT`, `BNBUSDT`, `SOLUSDT`, `XRPUSDT`, `ADAUSDT`, `DOGEUSDT`,
and `AVAXUSDT`, each with `5m`, `15m`, `1h`, and `4h` timeframes.

`GET /candles?symbol=BTCUSDT&timeframe=5m&from=<unix-ms>&to=<unix-ms>&limit=500`
returns closed candles from Postgres. Supported MVP timeframes are `5m`, `15m`,
`1h`, and `4h`; `limit` is bounded to 1-5000. Unsupported symbols/timeframes
are rejected. Production search also rejects pairs outside this catalog and
returns HTTP 422 when the requested range contains fewer than 21 candles.

## Authentication

`POST /auth/register` creates a bcrypt-backed user. `POST /auth/login` sets an
httpOnly SameSite=Lax `session` JWT cookie with a one-hour expiry. The configured
secret must contain at least 16 characters. All routes
except health/register/login require that cookie. Logout clears it.

## WebSocket

`GET /ws` sends tagged JSON messages on one authenticated connection:

```json
{"type":"CANDLE_UPDATE","payload":{}}
{"type":"TRADE_TICK","payload":{"symbol":"BTCUSDT","tradeId":1,"tradeTime":0,"price":0,"quantity":0,"side":"BUY"}}
{"type":"SEARCH_PROGRESS","payload":{"tested":1,"total":1}}
{"type":"LEADERBOARD_UPDATE","payload":[]}
```

Search is started through the REST endpoint in the current MVP; the WebSocket
is the server-push channel. A client selects only the market events it needs:

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
