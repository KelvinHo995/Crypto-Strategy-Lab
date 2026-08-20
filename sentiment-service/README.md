# sentiment-service

FastAPI service that scores news sentiment.

## Run
```bash
uv sync
uv run uvicorn app.main:app --reload --port 8000
```

## Structure
- app/main.py — routes
- app/schemas.py — request/response contracts (Pydantic)
- app/model.py — model loading + inference

## API
See /docs/contracts.md at repo root. /analyze shape is placeholder until day-1 sync.
