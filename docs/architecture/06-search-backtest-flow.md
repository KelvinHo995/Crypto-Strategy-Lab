# 6. Search → Backtest → Evaluate → Rank → Leaderboard

This is the continuous loop the spec calls the architectural core (ch.23,
ch.47): "Generate → Execute → Measure → Rank → Improve → Generate...".

## End-to-end flow

```
StrategyGenerator.Generate()
     │  produces a CandidateStrategy (see 05-strategy-flow.md)
     ▼
Queue.Enqueue(BacktestJob)          ← Queue interface; InMemoryQueue today,
     │                                 not Kafka (ADR-0004)
     ▼
Worker pool (N workers, each calls Queue.Dequeue)
     │  each worker: resolve strategies via Registry → run Backtester
     ▼
Backtester
     │  simulates trades over a fixed historical candle range, given
     │  pair/coin, test period (from–to), and starting capital as inputs
     │  → []Trade: { pair, entryTime, direction (LONG/SHORT), volumeUSD,
     │               entryPrice, stopLoss, takeProfit, exitPrice,
     │               transactionCost, slippage (simulated ~5bps), profit }
     ▼
Evaluator
     │  Return, Win Rate, Wins/Losses count, Max Drawdown, Trade Count,
     │  Total Profit, (Profit Factor, Sharpe)
     │  — deliberately a separate component from Backtester (ADR-0003):
     │    "did the strategy trade correctly" vs "was the outcome good" are
     │    different questions with different failure modes
     ▼
experiment.Result  { CandidateID, StrategyVersions (provenance), Return,
                      MDD, TradeCount, Status, CreatedAt }
     ▼
Ranking (Overall Score = weighted combination of Return/WinRate/RiskScore
          — team must state the exact formula, spec ch.21)
     ▼
Repository → Supabase (Postgres) `experiments` table
     ▼
Leaderboard (Top-K, K=10 by default)
     │  new candidate beats current #K → replaces it
     ▼
WebSocket → { "type": "LEADERBOARD_UPDATE", payload: Result[] }
     │        { "type": "SEARCH_PROGRESS", payload: {tested, total} }
     ▼
Frontend — leaderboard re-renders without a page reload (spec ch.33 step 9)
```

## Why a queue + worker pool, not a sequential loop

Spec ch.32.5 poses this directly: running 1,000+ backtests sequentially vs.
`Job Queue + Workers`. The queue/worker split is what lets the number of
workers scale independently of how candidates are generated or how results
are stored — see [ADR-0004](../adr/0004-inprocess-job-queue-not-kafka.md).
Scale math the team already worked out (PLAN.md, spec ch.43): at ~2s per
candidate on one worker, 10,000 candidates takes ~20,000s (>5.5h) serially;
the fix is more workers pulling from the same queue, not a faster single
worker.

## Why the queue is behind an interface

`Queue` (`Enqueue`/`Dequeue` over `BacktestJob`) is an interface;
`InMemoryQueue` is today's only implementation. Nothing else — the worker
pool, `StrategyGenerator`, `Backtester` — depends on `InMemoryQueue`
directly. This is the same Replaceability shape as the `StrategyGenerator`
swap below: a future `RedisQueue` or `KafkaQueue` (if distributed workers
ever become a real requirement) only has to satisfy the interface — see
[ADR-0004](../adr/0004-inprocess-job-queue-not-kafka.md) and
[07-quality-attributes.md](07-quality-attributes.md)'s Replaceability
section.

## Why the generator is swappable

`StrategyGenerator` is an interface:

```go
type StrategyGenerator interface { Generate() CandidateStrategy }
```

Today: `RandomGenerator`. Spec ch.42's scenario is: swap in
`DomainGuidedGenerator` (or later `GeneticGenerator`) without the Job Queue,
Backtester, Evaluator, or Leaderboard changing at all — everything downstream
only ever sees a `CandidateStrategy`, never how it was produced. This is the
same "consumer defines the interface, producer is swappable" shape as the
Strategy/Registry boundary in
[05-strategy-flow.md](05-strategy-flow.md).

## Stop condition

The loop must not be `while(true)` (spec ch.23, explicitly flagged as the
thing *not* to do). Configurable stop conditions: candidate count cap
(e.g. 100), wall-clock cap (e.g. 1 hour), or no-improvement-in-N-iterations
(e.g. 50). The team must pick and document which one(s) MVP supports.

## Trade simulation realism

Each simulated trade accounts for stop-loss/take-profit exit levels,
transaction cost (fee %), and slippage/spread (simulated ~5bps) — not just
an idealized fill at signal price. This isn't extra scope for its own sake:
Win Rate and Total Profit are only meaningful if a "win" is decided by
whether SL or TP was hit first, and cost/slippage are what stop the
Backtester from reporting profits that wouldn't survive contact with a real
order book. This stays inside `Backtester` (see
[ADR-0003](../adr/0003-separate-backtester-evaluator.md)) — the Evaluator
consumes the resulting `[]Trade`, it doesn't re-derive fills.

## Provenance

Every `Result` records `StrategyVersions` — which exact version of each
constituent strategy (and, if `SentimentStrategy` is included, which
sentiment model version) produced this result. Spec ch.36 calls this
Reproducibility: "Experiment #122 luôn biết chính xác nó đã sử dụng strategy
nào" — a leaderboard entry must be traceable to the exact code that produced
it, not just a strategy name that may have since changed behavior.
