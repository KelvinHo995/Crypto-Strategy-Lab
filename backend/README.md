# backend

Go API server — Market Data, Strategy, Experiment domains.

## Run
```bash
go run ./cmd/server
```

## Structure
- `cmd/server` — entrypoint, HTTP/WebSocket wiring only, no business logic
- `internal/market` — Candle, exchange adapters (owner: Person 1)
- `internal/strategy` — Strategy interface, registry, generators (owner: Person 2)
- `internal/experiment` — Backtester, Evaluator, weighted Ranking, Queue + 3-worker pool, Result provenance (owner: Person 3)

`POST /search/start` is asynchronous: a successful request returns HTTP 202
with `STARTED`; use `GET /experiments/{searchId}` to observe
`PENDING|RUNNING|COMPLETED|FAILED`.

## Contracts
See `/docs/contracts.md` at repo root. Types in this repo are placeholders until
day-1 sync — check that file before assuming a struct shape is final.
