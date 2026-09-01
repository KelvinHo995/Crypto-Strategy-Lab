# ADR-0013: Durable Postgres jobs and bounded local caches, not Redis/Kafka

**Status:** Accepted
**Owner:** Experiment / Market Data

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
- Use bounded process-local caches: 128 candle ranges for 60 seconds, a
  write-invalidated leaderboard snapshot for three seconds, and 4,096 sentiment
  timestamp scores for five minutes. Errors are not cached.
- Coalesce concurrent identical misses and use generation invalidation so an
  older in-flight read cannot repopulate stale state.

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
  cache, but correctness is preserved by TTL and invalidation.
- Replacing `PostgresQueue` with Redis Streams remains possible through the same
  queue contract. Kafka is still unjustified without replay or multiple consumer
  groups.
