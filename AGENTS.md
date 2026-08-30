# Repository Guidelines

## Project Structure & Module Organization

Crypto Strategy Lab has three main work areas:

- `backend/`: Go API server. Entrypoints are `cmd/server` and `cmd/backfill`; domain packages are under `internal/market`, `internal/strategy`, `internal/experiment`, `internal/auth`, and `internal/httpx`.
- `frontend/`: React + TypeScript + Vite app. Feature code is in `src/features/*`, shared utilities in `src/shared`, styling in `src/index.css` and `src/App.css`, and assets in `public/` or `src/assets/`.
- `sentiment-service/`: FastAPI sentiment API, with source in `app/` and tests in `tests/`.

Architecture notes, contracts, and ADRs are in `docs/`; database setup starts with `backend/migrations/0001_init.sql`.

## Build, Test, and Development Commands

- `cd backend && go run ./cmd/server`: start the Go API server.
- `cd backend && go run ./cmd/backfill`: backfill the fixed market dataset; safe to rerun.
- `cd backend && go test ./...`: run all Go tests.
- `cd frontend && npm run dev`: start the Vite development server.
- `cd frontend && npm run build`: type-check and build the frontend.
- `cd frontend && npm run lint`: run ESLint for TypeScript/React files.
- `cd sentiment-service && uv sync`: install Python dependencies.
- `cd sentiment-service && uv run uvicorn app.main:app --reload --port 8000`: run the sentiment API.
- `cd sentiment-service && uv run python -m unittest discover -s tests -v`: run sentiment-service tests.

## Coding Style & Naming Conventions

Use `gofmt` for Go changes. Keep package names short and lowercase, and place tests beside code as `*_test.go`. In the frontend, use TypeScript, React function components, PascalCase component files, and camelCase hooks/services. Follow the existing feature-folder pattern. Python code should target 3.11+, use Pydantic schemas, and keep API contracts stable.

## Testing Guidelines

Prefer focused unit tests near changed behavior. Go tests use the standard `testing` package and should cover touched packages. Frontend currently has lint/build verification rather than a test runner, so run both before UI PRs. Sentiment tests use `unittest` under `sentiment-service/tests`.

## Commit & Pull Request Guidelines

Recent history uses short conventional-style subjects such as `fix(market): ...`, `chore(go): ...`, and `docs: ...`, plus merge commits. Keep commits imperative and scoped when useful. PRs should summarize behavior changes, list verification commands, link issues/docs, and include screenshots for visible frontend changes. Note database, environment, or contract changes explicitly.

## Security & Configuration Tips

Do not commit secrets. The backend needs `DATABASE_URL` and a `JWT_SECRET` of at least 16 characters. Apply migrations before persistence flows, and keep contracts aligned with `PLAN.md` and `docs/architecture`.
