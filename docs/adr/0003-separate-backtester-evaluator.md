# ADR-0003: Backtester and Evaluator are separate components

**Status:** Accepted
**Owner:** Experiment (Person 3)

## Context

Spec ch.20 draws a line between "did the strategy generate the trades it
should have" (simulation correctness) and "was the outcome good"
(performance metrics) — explicitly warning against judging a strategy by
Profit alone (ch.20's Strategy A vs Strategy B example: +30% with a -45%
drawdown vs +25% with a -8% drawdown — profit alone picks the wrong winner).
Spec ch.20 states the underlying principle directly: "Strategy Evaluation
phải tách biệt khỏi Strategy Implementation."

## Decision

`Backtester` takes a strategy already constructed from a `CandidateStrategy`
plus a historical candle range and produces a trade list (entry/exit price,
timestamp, per-trade P/L) — it
answers "what happened." `Evaluator` takes that trade list and produces
metrics (Return, Win Rate, Max Drawdown, Trade Count, wins/losses, and Total
Profit) — it answers "was that good." Backtester never computes a ranking-relevant
score; Evaluator never simulates a trade.

```
CandidateStrategy → Registry/BuildFromCandidate → Strategy → Backtester → []Trade → Evaluator → Metrics
```

## Alternatives considered

- **One `runBacktest()` function that returns a Result directly** (fuses
  simulation and scoring). Rejected — changing the scoring formula (e.g.
  adding Sharpe Ratio, or reweighting how MDD affects Overall Score) would
  require touching the same function that owns trade simulation
  correctness, increasing the chance a scoring change introduces a
  simulation bug.
- **Evaluator as a frontend concern** (backend returns raw trades, browser
  computes metrics). Rejected — spec ch.44 explicitly lists "Frontend chứa
  business logic" (computing profit/ranking client-side) as an anti-pattern;
  metrics must be consistent regardless of which client renders them, and
  must be computable without a browser (e.g. for the ranking/leaderboard
  pipeline itself).

## Consequences

- **Positive:** Evaluator is independently testable against fixed trade
  fixtures — no need to run a real backtest to verify a Max Drawdown
  calculation is correct.
- **Positive:** future metrics such as Sharpe or Profit Factor are additive to
  Evaluator and need not change Backtester's trade-simulation logic. They are
  not implemented in the current `Metrics` contract.
- **Cost:** one more interface boundary and one more data shape (`[]Trade`)
  to keep stable between the two components — documented alongside the other
  cross-component contracts in `docs/contracts.md`.
