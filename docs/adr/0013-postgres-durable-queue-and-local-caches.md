# ADR-0013: Durable Postgres jobs and process-local caches, not Redis/Kafka

**Status:** Accepted
**Owner:** Experiment / Market Data / Sentiment

> **Implementation evolution (2026-09-07):** the durable queue decision is
> unchanged, but the cache details evolved. Candle ranges use a bounded
> 128-entry/60-second cache and leaderboard reads use a write-invalidated
> three-second snapshot. `sentiment.TimeLookup` currently uses a time-sorted,
> one-minute-TTL slice with no explicit entry cap; its 24-hour `maxAge` limits
> selection semantics, not cache size. The original 4,096-entry/five-minute
> sentiment target below was not implemented.

## Context

The original in-process queue was adequate for the first single-process MVP,
but runtime review exposed a crash window: the API saved `PENDING` and then
enqueued to a channel in two operations. A crash between them lost the job. A
restart also discarded queued work. Repeated candle ranges, sentiment timestamps
and leaderboard refreshes were issuing avoidable Supabase queries.

The application still has one primary database, one job type and low traffic.
Adding Redis plus Kafka/RabbitMQ would create operational dependencies without a
second consumer group or event-replay requirement.

## Decision

- Persist compact backtest jobs in `experiment_jobs` on the existing Postgres.
- Commit the PENDING result and job in one transaction.
- Claim through `FOR UPDATE SKIP LOCKED`, with attempt count, bounded retry,
  lease timestamp, worker heartbeat and stale-lease reclaim.
- Store only pair/timeframe/range/config/provenance in a message. Workers read
  candles from the authoritative repository; messages never embed the dataset.
- Keep the `Queue` interface and `InMemoryQueue` for isolated unit tests.
- Use process-local caches: 128 candle ranges for 60 seconds, a
  write-invalidated leaderboard snapshot for three seconds, and a one-minute
  time-sorted sentiment observation slice reused across candle timestamps.
  Errors are not cached.
- Coalesce concurrent identical candle/leaderboard misses and use generation
  invalidation so an older in-flight read cannot repopulate stale state.
  `TimeLookup` serializes refresh and lookup under one mutex; `Invalidate` is
  available, but current server-side sentiment writes do not invoke it.

## Consequences

- Jobs survive backend restarts and may be claimed safely by multiple instances.
- After three failed claims, both the diagnostic queue row and public experiment
  result become `FAILED`; a worker crash on its last lease is finalized by the
  next queue poll instead of leaving a permanent `RUNNING` result.
- Postgres remains the only operational state dependency and source of truth.
- Queue polling adds a small idle query load (at most once per worker per second).
- Failed queue rows remain available for diagnosis; a later retention job may
  archive/delete them if volume justifies it.
- Caches are per process. Multi-replica hit ratios are lower than a shared Redis
  cache. Candle/leaderboard correctness is protected by TTL and invalidation;
  sentiment observations written during a live cache window may become visible
  to lookup up to one minute later.
- Replacing `PostgresQueue` with Redis Streams remains possible through the same
  queue contract. Kafka is still unjustified without replay or multiple consumer
  groups.
