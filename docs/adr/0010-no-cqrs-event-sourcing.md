# ADR-0010: CQRS and Event Sourcing are not used

**Status:** Accepted
**Owner:** whole team (Experiment domain most directly affected)

## Context

The professor's deck walks through both patterns in detail as candidates
for the Experiment domain: CQRS would split a write side
(`RunBacktest → Experiment state machine → Result/Metrics/Events`) from a
read side (`LeaderboardView`, denormalized for display); Event Sourcing
would make `ExperimentCreated → CandidateAssigned → BacktestStarted →
BacktestCompleted → StrategyEvaluated → LeaderboardPromoted` the source of
truth, with current state derived by replay. Both are explicitly named in
the deck's Appendix I as an expected vấn đáp question ("Có cần CQRS + Event
Sourcing?"). PLAN.md §6 already made this call informally; this ADR is that
decision written up properly, since the deck's suggested ADR list expects
it as its own entry.

## Decision

Plain CRUD: `experiments` is the aggregate result table, with normalized
`experiment_trades` and durable `experiment_jobs` support tables, all accessed
through the same experiment repository/queue boundaries. There is no separate
write-model/read-model projection and no event log as the system of record.

## Alternatives considered

- **CQRS.** Rejected for current scope: the write shape (a `Result` row)
  and the read shape (a `LeaderboardView` row — Rank, Strategy, Return, MDD,
  Trades) are nearly identical, just sorted differently for display. CQRS
  earns its complexity when read and write models genuinely diverge in
  shape, or need to scale independently (e.g., many read replicas serving
  heavy dashboard traffic against a comparatively slow write path). Neither
  is true here — data volume is thin (thousands of rows, not millions), and
  read load is a single dashboard, not a high-QPS public API.
- **Event Sourcing.** Rejected for current scope: the benefits — audit
  trail, replay, temporal history, richer debugging — are real, but
  provenance (the actual requirement, spec ch.36) is already satisfied by
  storing `StrategyVersions`/`CandidateID`/`DatasetPeriod` directly on the
  `Result` row ([ADR-0009](0009-experiment-provenance-storage.md)), without
  needing to replay a full event history to answer "what produced this
  result?" The costs — event schema evolution over time, storage growth,
  replay/projection machinery — aren't justified when state-only storage
  already answers every question currently asked of it.

## Consequences

- **Positive:** a small state-oriented persistence model for the Experiment
  domain, easy to explain and defend without a CQRS projection pipeline or
  event replay machinery.
- **Positive:** consistent with the team's stated philosophy elsewhere
  (PLAN.md §6's "no driver forces this yet" reasoning for Redis, Kafka,
  Kubernetes) — this isn't an isolated simplicity call, it's the same
  judgment applied again.
- **Cost / explicit limit:** if a future requirement needs a true audit
  trail (every intermediate state transition an experiment went through,
  not just its final outcome) or read/write loads diverge sharply, this
  decision should be revisited. The `Repository` interface boundary
  (already used to keep persistence swappable) is what would make that
  migration a contained change rather than a rewrite — same reasoning as
  the `Queue` interface in [ADR-0004](0004-inprocess-job-queue-not-kafka.md).
