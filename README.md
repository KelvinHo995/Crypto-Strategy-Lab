# Crypto Strategy Lab

Full-stack quantitative strategy lab for realtime Binance candles, composable trading strategies, reproducible backtests, experiment ranking, and news sentiment analysis.

## Stack

- Go backend (`backend/`) — REST, authenticated WebSocket, market ingestion, strategies, search, backtesting, ranking, PostgreSQL persistence.
- React + TypeScript frontend (`frontend/`) — financial workstation UI with live API/WebSocket integration and explicitly labelled offline fallbacks.
- Python FastAPI sentiment service (`sentiment-service/`) — deterministic `crypto-lexicon/v2` model behind a replaceable HTTP contract.
- PostgreSQL/Supabase — users, candles, experiments, and sentiment observations.

## Quick start

1. Copy `backend/.env.example` to `backend/.env` and set `DATABASE_URL` plus a `JWT_SECRET` of at least 16 characters.
2. Set `DATABASE_URL`, then run `go run ./cmd/migrate` from `backend/`. It applies
   every idempotent migration in numeric order (`0001` through `0014`).
3. Start the three processes in separate terminals:

```bash
cd sentiment-service
uv sync
uv run uvicorn app.main:app --reload --port 8000
```

```bash
cd backend
go run ./cmd/server
```

```bash
cd frontend
pnpm install --frozen-lockfile
pnpm dev
```

Open `http://127.0.0.1:5173`. Before production searches, backfill the supported
markets. The default remains BTCUSDT/two years; a multi-market run can be scoped:

```bash
BACKFILL_SYMBOLS=BTCUSDT,ETHUSDT,BNBUSDT,SOLUSDT,XRPUSDT,ADAUSDT,DOGEUSDT,AVAXUSDT BACKFILL_DAYS=180 go run ./cmd/backfill
```

PowerShell users set those values through `$env:BACKFILL_SYMBOLS` and
`$env:BACKFILL_DAYS` before invoking the command. Re-running is safe because
candles are upserted by `(symbol,timeframe,open_time)`.

## Verification

```bash
cd backend && go test ./... && go vet ./...
cd frontend && pnpm run lint && pnpm run build
cd sentiment-service && uv run python -m unittest discover -s tests -v
```

Architecture decisions and runtime contracts live in `docs/`. See `PLAN.md` for
scope and ownership, and [`docs/e2e-testing.md`](docs/e2e-testing.md) for the
complete Supabase/Binance/fullstack verification runbook.

In `Live infrastructure` mode the Backtests screen never substitutes fixture
experiments or generated trades when an API call fails or returns no rows.
Offline samples are available only in the explicitly selected `DEMO` mode.

## Demo

See [`docs/e2e-testing.md`](docs/e2e-testing.md) §5 "Browser acceptance
scenarios" for the walkthrough: realtime multi-timeframe charts, strategy
selection, composite generation, backtest run, leaderboard, and sentiment
analysis, all against the real backend/Supabase.

## Team ownership and contributions

Primary ownership identifies the module each member is responsible for. Actual
implementation work also crosses module boundaries, as summarized below.

| Member | Primary ownership | Additional contributions |
|---|---|---|
| Võ Thành Đạt | Market Data + Sentiment Service | Binance resilience, RSS ingestion, sentiment integration and experiment provenance |
| **Vũ Thành Đạt (`yuno1310`)** | **Strategy + Search** | Strategy plugins and extensibility; random/continuous search; backend API, auth and runtime integration; multi-market streaming and backfill; durable PostgreSQL queue, bounded caches, worker metrics and performance evidence; frontend REST/WebSocket integration, multi-coin charts, sentiment and search-loop UI, responsive workstation and dark theme; experiment/leaderboard integrity; CI, integration/E2E tests, ADRs, contracts and operational documentation |
| Hồ Tấn Quốc | Experiment (Backtest/Evaluate/Rank) | Trade persistence and visualization, search-loop integrity, short-selling and runtime hardening |
| Trịnh Hạnh | Frontend | Cross-tab state, backtest payload synchronization, news pipeline UI and strategy presets |

Never commit `.env`, database credentials, JWT secrets, or exported cookie files.
