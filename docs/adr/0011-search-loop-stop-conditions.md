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
- **Cost:** multiple simultaneously-active stop conditions (e.g., "stop at
  either 1000 candidates OR 1 hour, whichever comes first") need an explicit
  precedence/combination rule when implemented — not yet specified, must be
  decided before the Continuous Loop milestone (M5) is built, not left
  implicit.
- **Open item:** job failure *retry* policy — does a `FAILED` `BacktestJob`
  get retried, how many times, before counting as a permanent failure? Not
  yet decided; a follow-up note (or ADR, if the answer turns out to be
  non-trivial) is needed once the worker pool is actually implemented.
