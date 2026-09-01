# ADR-0004: In-process job queue behind a `Queue` interface, not Kafka/RabbitMQ

**Status:** Superseded by [ADR-0013](0013-postgres-durable-queue-and-local-caches.md)
**Owner:** Strategy + Search (Person 2), Experiment (Person 3)

## Context

Spec ch.24 warns against a naive sequential loop over candidates
(`for 100000 strategies: calculate → backtest → save → update UI` all in one
function) and asks for a design that supports concurrent workers, retry on
worker failure, pause/resume, and later scaling (ch.32.5, ch.43). The
question is *how* to get concurrency — a real message broker (Kafka,
RabbitMQ) is one way; Go's built-in channels + goroutines are another.

A second, related question came up after the fact: if we start with an
in-process channel, can we swap to Redis or Kafka later *without* rewriting
the worker pool, the `StrategyGenerator`, or the `Backtester`? That's a
concrete **Replaceability** test (the same category of question the
project's rubric asks about `StrategyGenerator`: "does swapping the
implementation touch anything downstream?").

## Decision

Define a `Queue` interface and a `BacktestJob` contract in `internal/experiment`,
and implement it today with an in-memory, channel-backed queue:

```go
type BacktestJob struct {
    ID         string
    Candidate  strategy.CandidateStrategy
    EnqueuedAt int64
}

type Queue interface {
    Enqueue(ctx context.Context, job BacktestJob) error
    Dequeue(ctx context.Context) (BacktestJob, error)
}

// InMemoryQueue — the only implementation for MVP, wraps a buffered channel.
type InMemoryQueue struct {
    ch chan BacktestJob
}
```

The worker pool, the `StrategyGenerator`, and the `Backtester` all depend on
the `Queue` interface, never on `InMemoryQueue` directly. `InMemoryQueue` is
just today's implementation, injected at composition-root time
(`cmd/server`), same pattern as `Repository` (PLAN.md §6: "DI qua interface —
Có dùng").

## Alternatives considered

- **Kafka / RabbitMQ.** Rejected for current scope — PLAN.md §6 states the
  reasoning directly: "Không có nhiều service cần decouple qua broker thật."
  There is exactly one producer (the generator loop) and one consumer group
  (backtest workers) within a single process; a broker's value — durability
  across process restarts, multiple independent consumer groups, cross-service
  decoupling — isn't needed when producer and consumer share a process and
  jobs are regenerable (a lost in-flight candidate can simply be
  regenerated, it's not an unrecoverable event like a lost financial
  transaction would be).
- **Sequential loop, no concurrency.** Rejected — spec's own scale math
  (ch.43): 1 worker at 2s/candidate × 10,000 candidates = 20,000s (>5.5h)
  serially. Doesn't meet the "100 → 100,000 candidates" scalability driver
  (ch.32.2).
- **Redis-backed queue.** Rejected alongside Kafka for the same reason —
  single process, single instance, no need for a shared queue visible to
  multiple processes (PLAN.md §6: "Traffic thấp, 1 instance, không cần
  shared cache giữa nhiều replica"). Kept as the most likely future
  `Queue` implementation if that changes (see Consequences).
- **Skip the interface, call the channel directly from the worker pool.**
  Rejected — this was the original version of this ADR. Costs one small
  interface definition; buys a documented, testable seam for the exact
  Replaceability question the project is graded on. Given the project
  already pays this cost for `Strategy`, `StrategyGenerator`, and
  `Repository`, doing it for the queue too is the consistent call, not an
  extra one.

## Consequences

- **Positive:** worker count is a config value, not an architecture change —
  scaling from 1 to N workers only changes a pool size.
- **Positive:** no operational dependency (broker to run, monitor, and keep
  available) for the 2-week build and demo.
- **Positive — Replaceability:** a future `RedisQueue` or `KafkaQueue` only
  needs to implement `Enqueue`/`Dequeue` against `BacktestJob`. The worker
  pool, `StrategyGenerator`, and `Backtester` don't change — this is now a
  concrete, demonstrable answer to "how do you know your architecture
  supports swapping the queue," the same shape as the `StrategyGenerator`
  Random→Genetic swap test (spec ch.42).
- **Cost / explicit limit:** `InMemoryQueue` does **not** scale across
  multiple machines — all workers share one process's memory and one queue.
  If a future requirement needs distributed workers (multiple backend
  instances sharing one candidate queue), swap the `Queue` implementation —
  the interface is what makes that a contained change instead of a rewrite.
- **Cost:** job durability is best-effort — a process crash mid-run loses
  in-flight (not yet completed) jobs. Accepted because jobs are cheap to
  regenerate and idempotent to rerun (spec ch.43's own framing of this
  tradeoff), unlike, say, a payment.
- **Cost:** `BacktestJob.EnqueuedAt` and `ID` exist specifically to support
  observability (queue latency, per-job tracking) — see
  [ADR-0011](0011-search-loop-stop-conditions.md) — which the interface
  wouldn't need if it only carried a bare `CandidateStrategy`.
