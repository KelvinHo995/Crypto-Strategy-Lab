# Crypto Strategy Lab

## Stack
- Backend: Go 1.22 (backend/)
- Frontend: React + TypeScript, pnpm (frontend/)
- Sentiment Service: Python 3.11, FastAPI, uv (sentiment-service/)

## Structure
- backend/cmd/server — Go entrypoint
- backend/internal/{market,strategy,experiment} — domain packages, one owner each
- frontend/src/features/{market,strategy,experiment,news} — feature-based FE structure
- sentiment-service/app/ — FastAPI service
- docs/adr/ — Architecture Decision Records
- docs/contracts.md — cross-language API/WS schemas (pending day-1 meeting)

## Status
Contracts (Candle, Strategy interface, ExperimentResult, Sentiment API shape) are
scaffolded with placeholder types. Not finalized yet — first team sync will lock
these down. Do not build features against placeholder shapes assuming they're final.

## Quickstart

# backend
```bash
cd backend && go run ./cmd/server
```

# frontend
```bash
cd frontend && pnpm install && pnpm dev
```

# sentiment-service
```bash
cd sentiment-service && uv sync && uv run uvicorn app.main:app --reload --port 8000
```

## Team & Ownership
| Person | Owns |
|---|---|
| Person 1 | Market Data |
| Person 2 | Strategy + Search |
| Person 3 | Experiment (Backtest/Evaluate/Rank) |
| Person 4 | Frontend |

## Working rules
- Branch naming: feat/<domain>-<short-desc>, e.g. feat/market-binance-adapter
- Keep branches short-lived (1-2 days max), PR into main frequently, small diffs
- 1 approval required before merge to main
- Each person owns their domain folder — cross-domain changes need that person's review
- Type-check before pushing: go vet ./... (backend), pyright (sentiment-service),
  tsc --noEmit (frontend) — don't rely on catching drift at integration time
