# frontend

React + TypeScript dashboard.

## Run
```bash
pnpm install
pnpm dev
```

## Structure
Feature-based, one folder per backend domain:
- src/features/market — chart, candle display
- src/features/strategy — strategy picker/config
- src/features/experiment — leaderboard, backtest results
- src/features/news — news + sentiment panel
- src/shared/api — REST + WebSocket client
- src/shared/components — dumb/presentational UI only, no domain logic

## API
See /docs/contracts.md at repo root for WS/REST message shapes.
