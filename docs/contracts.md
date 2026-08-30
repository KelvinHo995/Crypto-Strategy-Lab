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
`GET /experiments/{id}` returns one provenance snapshot.

## Authentication

`POST /auth/register` creates a bcrypt-backed user. `POST /auth/login` sets an
httpOnly SameSite=Lax `session` JWT cookie with a one-hour expiry. All routes
except health/register/login require that cookie. Logout clears it.

## WebSocket

`GET /ws` sends tagged JSON messages on one authenticated connection:

```json
{"type":"CANDLE_UPDATE","payload":{}}
{"type":"SEARCH_PROGRESS","payload":{"tested":1,"total":1}}
{"type":"LEADERBOARD_UPDATE","payload":[]}
```

Search is started through the REST endpoint in the current MVP; the WebSocket
is the server-push channel.

## Sentiment

`POST /analyze` accepts `{newsId,text}` and returns
`{newsId,sentiment,score,model:{name,version},createdAt}`. The Go client treats
service errors explicitly; `SentimentStrategy` degrades to its base strategy.
