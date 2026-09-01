# 6. Search → Backtest → Evaluate → Rank → Leaderboard

This is the continuous loop the spec calls the architectural core (ch.23,
ch.47): "Generate → Execute → Measure → Rank → Improve → Generate...".

## End-to-end flow

```
StrategyGenerator.Generate()
     │  produces a CandidateStrategy (see 05-strategy-flow.md)
     ▼
PostgresQueue.EnqueuePending(job, result)
     │  one transaction: PENDING provenance + compact durable job
     ▼
Worker pool (3 workers, SKIP LOCKED claim + lease heartbeat)
     │  load referenced candle range through bounded cache
     │  resolve strategies via Registry → run Backtester → Ack/Nack
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
Ranking (Score = 0.50×Return + 0.30×WinRate − 0.20×MDD)
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
worker. The original in-process decision is recorded in ADR-0004; the current
durability driver and implementation are recorded in
[ADR-0013](../adr/0013-postgres-durable-queue-and-local-caches.md).

## Why the queue is behind an interface

`Queue` (`Enqueue`/`Dequeue`/`Ack`/`Nack`) remains an interface.
`PostgresQueue` is production and `InMemoryQueue` remains a test implementation;
the worker pool and Backtester depend on neither concrete type. A future Redis
Streams implementation can replace the transport without changing evaluation
logic — see ADR-0013 and
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

The current HTTP slice accepts one candidate per request and therefore has an
implicit candidate-count cap of one; it never uses `while(true)`. The next
Search integration will batch generated candidates and use the configured
candidate-count cap as the MVP stop condition. Wall-clock and no-improvement
limits remain extension points rather than partially implemented behaviour.

## Implemented runtime states

`POST /search/start` validates the named strategies, snapshots candidate and
version provenance, saves `PENDING`, enqueues the job, and returns HTTP 202
with `{searchId, status:"STARTED"}`. The initial row and queue message commit
atomically, so a crash cannot leave an unqueued PENDING result. A worker changes the same result to
`RUNNING`, executes Backtester and Evaluator, then persists `COMPLETED`; a
candidate-resolution failure becomes `FAILED`. Infrastructure writes retry up
to three claims with bounded backoff; stale leases are reclaimable. Exhausting
the third claim atomically marks both the job and experiment `FAILED`. Workers
heartbeat both the job lease and experiment `updated_at`. Each result stores
`searchId`/`searchTotal` (`id`/`1` for today's single-candidate request) so a
future bounded batch can group candidates without another contract change. Repository reads use
an indexed, cached Top-100 snapshot; incomplete jobs remain visible below
completed results.

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

Built-in strategies currently use the explicit release label `v1`. A future
Sentiment candidate must replace/add the actual model name/version received
from the Sentiment API before enqueueing; the worker persists the supplied
snapshot unchanged.
