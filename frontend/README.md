# frontend

React/Vite client for Crypto Strategy Lab.

## Development

Start the Go backend on port 8080 first, with migrations and backfill completed.
Then run:

```bash
pnpm install --frozen-lockfile
pnpm dev
```

### Visual system

The application uses a light financial-workstation theme: neutral white/slate surfaces, cobalt navigation and controls, emerald/red reserved for positive/negative market semantics, and a responsive sidebar that becomes a compact navigation row on smaller screens. Decorative emoji are intentionally avoided; interface actions use the shared Lucide icon set or plain text.

The unauthenticated screen offers an explicit offline demo. API-backed screens continue to label mock fallbacks so simulated values cannot be mistaken for production market data.

Feature workspaces are loaded on demand with React `lazy`/`Suspense`. The production build emits separate Market, Strategy, Backtest, and News chunks so the initial application bundle stays below the Vite 500 kB warning threshold.

Vite proxies `/auth`, `/candles`, `/strategies`, `/search`, `/experiments`,
`/sentiment`, and `/ws` to the backend, keeping the JWT cookie and WebSocket same-origin. For a
production deployment, serve both behind the same origin or set
`VITE_API_BASE_URL` and `VITE_WS_HOST` with a corresponding trusted proxy.

The app authenticates before opening WebSocket. Market history, realtime candles,
strategy registry, search and leaderboard use backend data. Market charts fall
back to visibly labelled mock candles when the backend/database is unavailable.
The News screen can send a sample article through the real Go → FastAPI →
PostgreSQL sentiment path and labels the result `LIVE`. The article feed,
crawler/extraction visuals, 24-hour aggregate, and per-trade details remain
labelled demo data because production collector/trade-detail endpoints are
outside the current MVP.

When the backend is intentionally unavailable, choose **Chạy UI demo offline**.
Mock-backed panels are labelled `MOCK` or `DEMO` so they cannot be mistaken for
production data.
