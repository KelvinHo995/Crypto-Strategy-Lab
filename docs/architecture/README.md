# Architecture Documentation

Trimmed arc42-style breakdown — one file per concern, sized to what this project's
grading criteria actually asks for (see PDF spec ch.45 "Deliverables" and ch.32
"Architectural drivers"). Full arc42 has 12 sections; we only need these 7.

| File | Answers |
|---|---|
| [01-context.md](01-context.md) | What is this system, who uses it, what's outside its boundary? |
| [02-modules.md](02-modules.md) | What are the deployable units and internal packages, who owns what? |
| [03-data-flow.md](03-data-flow.md) | How does historical data get from Binance into a backtest? |
| [04-realtime-flow.md](04-realtime-flow.md) | How does a live price tick reach the browser? |
| [05-strategy-flow.md](05-strategy-flow.md) | How does a strategy plug in, and how do strategies combine? |
| [06-search-backtest-flow.md](06-search-backtest-flow.md) | How does search → backtest → evaluate → rank → leaderboard work? |
| [07-quality-attributes.md](07-quality-attributes.md) | What non-functional pressures shaped these decisions, and what must we be able to answer at defense (vấn đáp)? |

Decisions with real alternatives and consequences live in [`../adr/`](../adr/README.md)
as individual ADRs, not here — this folder describes the shape of the system as
agreed; the ADRs argue for *why* it has that shape.

Source of truth for cross-language contracts (Candle, Strategy interface,
ExperimentResult, Sentiment API, WebSocket messages) is [contracts.md](../contracts.md)
(referenced by the root and backend READMEs). These architecture docs describe
flow and ownership, not wire formats — don't duplicate struct definitions here,
link to the contract instead.

The implementation delta and its verified project impact are recorded in
[backend-implementation-review.md](../backend-implementation-review.md).
