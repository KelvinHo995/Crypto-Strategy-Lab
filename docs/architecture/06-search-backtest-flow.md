# 6. Search → Backtest → Evaluate → Rank → Leaderboard

This is the continuous loop the spec calls the architectural core (ch.23,
ch.47): "Generate → Execute → Measure → Rank → Improve → Generate...".

## End-to-end flow

```
StrategyGenerator.Generate()
     │  produces a CandidateStrategy (see 05-strategy-flow.md)
     ▼
Queue submission
     │  /search/start → PostgresQueue.EnqueuePending(job, result), one transaction
     │  /search/loop  → Repository.Save then Queue.Enqueue (open atomicity gap)
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
experiment.Result  { SearchID, SearchTotal, CandidateID,
                      StrategyVersions (provenance), Return,
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
Frontend — Discovery progress and leaderboard re-render only from server events
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

`POST /search/start` remains the explicit one-candidate path. The Discovery UI
uses `POST /search/loop`, which validates `maxCandidates` (2–200), optional
`maxDurationSeconds` (`0` or 60–3600), and optional non-negative
`noImprovementLimit`. Enabled conditions have OR semantics: reaching any one
stops further generation; queued work continues.

That describes the intended policy, but there are implementation caveats that
must remain visible until the backend work is complete:

- wall-clock time currently measures the generator goroutine, not the complete
  search run;
- the production PostgreSQL queue is durable but effectively unbounded, so the
  generator can enqueue the requested maximum before completion-driven
  no-improvement state changes;
- the observer is removed when generation returns, while queued jobs can still
  be running;
- there is no persisted search-run status/reason, so an early stop does not emit
  a terminal `STOPPED` event and progress can remain below the requested total;
- candidate dedup retries ten times, then may still enqueue a duplicate.

Consequently max-candidates is the only currently deterministic external bound.
The other controls are accepted inputs but are not release-complete semantics.

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
`searchId`/`searchTotal` (`id`/`1` for the single-candidate request). Loop jobs
use one shared `searchId` and the requested maximum as `searchTotal`. Repository reads use
an indexed, cached Top-100 snapshot; incomplete jobs remain visible below
completed results.

`POST /search/loop` returns HTTP 202 with
`{searchId,status:"STARTED",maxCandidates}` after validating that at least 21
candles exist. It does not wait for generated jobs. Today there is no standalone
search-run resource: `STARTED` is an acknowledgement, not a queryable lifecycle
record. Per-candidate states remain `PENDING|RUNNING|COMPLETED|FAILED`.

`SEARCH_PROGRESS` currently contains only `{tested,total}` and is global to an
authenticated WebSocket connection. `tested` counts terminal candidate rows;
`total` is the requested maximum. The frontend never advances this value
locally. It treats `tested == total` as `COMPLETED`, reports REST failures as
`FAILED`, and supports additive future `{searchId,status,reason}` fields for
explicit `STOPPED|FAILED` search-run events. Until those fields exist, an early
stop cannot be represented accurately in the UI.

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
