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
- `internal/experiment` — Backtester, Evaluator, Ranking (owner: Person 3)

## Contracts
See `/docs/contracts.md` at repo root. Types in this repo are placeholders until
day-1 sync — check that file before assuming a struct shape is final.
