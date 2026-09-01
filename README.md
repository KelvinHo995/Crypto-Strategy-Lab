# Crypto Strategy Lab

Full-stack quantitative strategy lab for realtime Binance candles, composable trading strategies, reproducible backtests, experiment ranking, and news sentiment analysis.

## Stack

- Go backend (`backend/`) — REST, authenticated WebSocket, market ingestion, strategies, search, backtesting, ranking, PostgreSQL persistence.
- React + TypeScript frontend (`frontend/`) — financial workstation UI with live API/WebSocket integration and explicitly labelled offline fallbacks.
- Python FastAPI sentiment service (`sentiment-service/`) — deterministic `crypto-lexicon/v1` model behind a replaceable HTTP contract.
- PostgreSQL/Supabase — users, candles, experiments, and sentiment observations.

## Quick start

1. Copy `backend/.env.example` to `backend/.env` and set `DATABASE_URL` plus a `JWT_SECRET` of at least 16 characters.
2. Apply `backend/migrations/0001_init.sql` and `backend/migrations/0002_sentiment_results.sql` to PostgreSQL.
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

Open `http://127.0.0.1:5173`. Run `go run ./cmd/backfill` from `backend/` before production searches so the selected range contains enough closed candles.

## Verification

```bash
cd backend && go test ./... && go vet ./...
cd frontend && pnpm run lint && pnpm run build
cd sentiment-service && uv run python -m unittest discover -s tests -v
```

Architecture decisions and runtime contracts live in `docs/`. See `PLAN.md` for scope and ownership.

## Team ownership

| Member | Area |
|---|---|
| Võ Thành Đạt | Market Data + Sentiment Service |
| Vũ Thành Đạt | Strategy + Search |
| Hồ Tấn Quốc | Experiment (Backtest/Evaluate/Rank) |
| Trịnh Hạnh | Frontend |

Never commit `.env`, database credentials, JWT secrets, or exported cookie files.
