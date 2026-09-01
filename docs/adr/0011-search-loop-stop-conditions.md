# ADR-0011: Stop conditions and observability of the Search Loop

**Status:** Accepted
**Owner:** Experiment (Person 3), Strategy + Search (Person 2)

## Context

Spec ch.23 explicitly warns against `while(true)` as the continuous strategy
loop's termination condition — an uncontrolled loop with no visibility into
its own progress. The deck's rubric (Appendix C) asks directly whether the
system can answer "queue depth, job latency, failure count" — i.e., whether
the loop's *state while running* is observable, not just its final leaderboard
output.

## Decision

**Stop conditions** are configuration for a given search run, not hardcoded,
and can be combined:
- max candidate count (e.g., stop after 100 candidates)
- max wall-clock time (e.g., stop after 1 hour)
- no-improvement-in-N-iterations (e.g., stop if best score hasn't improved
  in 50 iterations)
- user-cancel (explicit stop signal from the frontend)

**Observability** is carried by two existing mechanisms rather than a new
subsystem:
- `SEARCH_PROGRESS` WebSocket messages (`{tested, total}`, PLAN.md §3.5),
  extended to also carry current best score and failure count.
- `BacktestJob.Status` (`PENDING|RUNNING|COMPLETED|FAILED`, per the
  `Result` schema) and `BacktestJob.EnqueuedAt`
  ([ADR-0004](0004-inprocess-job-queue-not-kafka.md)) give per-job status
  and let queue latency (time between enqueue and dequeue) be measured
  directly, without a separate metrics store.

**MVP implementation status (2026-08-27):** one accepted HTTP request creates
one bounded job, so the running backend cannot enter an uncontrolled loop.
The queue and three-worker pool implement the per-job states above. Batched
candidate generation, user cancellation, and `SEARCH_PROGRESS` WebSocket
broadcasting remain integration work; they are not claimed as implemented.
For the first batched version, max candidate count is the required stop
condition and wins when any additional optional limit is reached first.

## Alternatives considered

- **Single hardcoded stop condition** (always exactly N candidates).
  Rejected — a quick sanity-check run during development and an
  overnight/demo search need different limits; hardcoding one forces a code
  change for every different need.
- **No progress observability, poll final leaderboard state only.**
  Rejected — this is exactly the failure mode spec ch.23 warns about: a
  stuck loop, a silently-failing worker, and a slow-but-healthy loop all
  look identical from the outside if the only signal is "has the
  leaderboard changed yet."
- **A dedicated metrics/observability service** (e.g., push queue depth and
  job latency to a separate monitoring stack). Rejected for MVP — no driver
  currently requires it; `SEARCH_PROGRESS` plus `BacktestJob.Status` answers
  the rubric's verification question without adding a new component.

## Consequences

- **Positive:** stop conditions are configuration, not code — changing
  search run limits doesn't need a rebuild.
- **Positive:** `SEARCH_PROGRESS` gives the frontend (and the graded demo)
  a live, honest signal of loop health, directly answering the deck's
  Observability rubric question.
- **Rule:** combined stop conditions use OR semantics: the first limit reached
  stops further generation. Already-enqueued jobs finish normally.
- **Rule:** MVP workers do not retry failed jobs. A deterministic bad
  candidate would fail identically, while automatic retry could duplicate
  compute invisibly. Manual resubmission creates a new traceable job ID.
- **Rule:** a `RUNNING` result whose worker or process crashed mid-job would
  otherwise sit stuck forever with nothing to notice it. A periodic sweep
  (`experiment.Sweeper`, on a 1-minute tick) fails any `RUNNING` result whose
  `updated_at` hasn't moved in 15 minutes, via one atomic SQL statement
  (`Repository.MarkStaleRunningFailed`) — it never touches `Queue`, so
  swapping `InMemoryQueue` for `RedisQueue`/`KafkaQueue` later needs no
  change here. Same manual-resubmission path as any other failure.
- **Rule:** `MarkStaleRunningFailed` is guarded by a Postgres advisory
  transaction lock (`pg_try_advisory_xact_lock`), so multiple server
  instances can call it concurrently without duplicating work or racing —
  whichever instance gets the lock does the sweep, everyone else sees it's
  held and skips that tick. Chosen over an in-memory per-process timer
  because a purely in-process signal for "how long has this been running"
  can't be compared across instances, and over moving the sweep out of the
  app entirely (e.g. a Supabase Edge Function/`pg_cron` job) because that
  would give Postgres responsibilities beyond the "just a host" role
  ADR-0012 deliberately scoped it to, for no correctness benefit here.
