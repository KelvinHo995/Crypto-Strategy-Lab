# sentiment-service

FastAPI sentiment service with a deterministic, versioned MVP model
(`crypto-lexicon/v2`). It tokenizes headline/body vocabulary and is replaceable
behind the same `/analyze` contract.

```bash
uv sync
uv run uvicorn app.main:app --reload --port 8000
uv run python -m unittest discover -s tests -v
```

Endpoints: `GET /health` and `POST /analyze` with `{newsId,text}`.
