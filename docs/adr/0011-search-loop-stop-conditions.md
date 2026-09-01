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

**Implementation status (2026-09-01):** `POST /search/loop` generates
candidates via `strategy.RandomGenerator` and enqueues them one at a time
(`experiment.RunSearchLoop`) until the first stop condition trips — OR
semantics, exactly as decided above:
- **max candidate count** — required (`MaxCandidates`, validated 2–200),
  always the backstop even if nothing else is set.
- **max wall-clock time** — optional (`MaxDurationSeconds`, validated
  60–3600 if set).
- **no-improvement-in-N** — optional (`NoImprovementLimit`). Tracked
  event-driven via a new `WorkerPool.AddObserver` (multiple observers, not
  just the one `SetObserver` hook the WebSocket hub already used) rather
  than polling the database before every candidate — the loop's observer
  gets pushed a notification the instant one of its own candidates
  completes, updates an in-memory best-score/steps-since-improvement pair,
  and the generator checks that before producing the next candidate. Two
  honest imprecisions, both accepted:
  - it counts *generation attempts*, not *completions in generation
    order* — 3 workers finish out of order, so "N iterations" means "N
    times the generator checked and found no new best," not "N
    consecutive backtests."
  - because the queue lets the generator race ahead of what's actually
    finished (up to the queue's buffer size), the stop can overshoot
    `NoImprovementLimit` by a few candidates before it lands. Tightening
    this would mean waiting for each candidate to finish before generating
    the next, serializing generation to completion pace and starving the
    3-worker pool's parallelism — not worth it for a bound that's already
    naturally capped by queue depth.
- **user-cancel** — still deferred. Nothing on the frontend calls a cancel
  endpoint today, and the other three conditions already guarantee
  termination (no `while(true)` risk) without it.

Within one loop run, generated candidates are deduplicated locally (a
`seen` set keyed by sorted strategies + params + policy, in-memory, scoped
to that single run) — cheap and simple, even though the actual param space
(see `strategy.RandomGenerator`) is large enough that accidental collisions
are already very unlikely at the validated candidate-count range.

`SEARCH_PROGRESS` (`{tested, total}`) is now real, not a placeholder:
`Result`/`BacktestJob` carry `SearchID`+`SearchTotal`, `Repository.List`
gained `ListBySearch`, and `Hub.JobUpdated` computes `tested` by counting
actual `COMPLETED`/`FAILED` rows for that search rather than hardcoding
`{0,1}`/`{1,1}`. This also unified the single-candidate `/search/start`
path onto the same mechanism (`SearchID = its own ID`, `SearchTotal = 1`)
instead of a separate, divergent code path.

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
- **Positive — Extensibility:** `WorkerPool.AddObserver` turned a single
  hardcoded callback into a small pub-sub registry — any new consumer
  (a failure counter for the deck's "how many job lỗi?" observability
  question, a per-search best-result cache, a future notification hook)
  is just another `AddObserver` call, with zero changes to `WorkerPool` or
  any existing observer. Same category of evidence as the `Queue`
  interface's Replaceability argument in ADR-0004, this time for
  Extensibility.
- **Rule (updated 2026-09-01):** deterministic candidate/build failures are
  terminal — retrying a bad candidate would just fail identically. Infrastructure
  failures while loading or persisting are Nack'ed and may be claimed up to
  three times with bounded backoff (`PostgresQueue`) under the same traceable
  job ID before becoming terminal too. Manual resubmission after that creates
  a new job ID, same as before — this supersedes the original "no automatic
  retry" rule, which held for the simpler `InMemoryQueue`-only MVP.
- **Rule:** a `RUNNING` result whose worker or process crashed mid-job would
  otherwise sit stuck forever with nothing to notice it. A periodic sweep
  (`experiment.Sweeper`, on a 1-minute tick) fails any `RUNNING` result whose
  `updated_at` hasn't moved in 15 minutes, via one atomic SQL statement
  (`Repository.MarkStaleRunningFailed`). `PostgresQueue` workers heartbeat the
  same `updated_at` while renewing their lease, so live work is not swept.
  See ADR-0013.
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
