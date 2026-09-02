# ADR-0011: Stop conditions and observability of the Search Loop

**Status:** Accepted
**Owner:** Experiment (Person 3), Strategy + Search (Person 2)

> **Implementation review (2026-09-02): Partial.** The endpoint and frontend
> integration exist, but terminal search-run state, atomic loop enqueue and
> reliable duration/no-improvement semantics are still open. This ADR records
> the intended decision and explicitly separates it from shipped behavior.

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
- `SEARCH_PROGRESS` WebSocket messages (`{tested, total}`, PLAN.md §3.5).
  `searchId`, terminal `status`, `reason`, best score and failure count are
  additive target fields; the current backend does not publish them yet.
- `BacktestJob.Status` (`PENDING|RUNNING|COMPLETED|FAILED`, per the
  `Result` schema) and `BacktestJob.EnqueuedAt`
  ([ADR-0004](0004-inprocess-job-queue-not-kafka.md)) give per-job status
  and let queue latency (time between enqueue and dequeue) be measured
  directly, without a separate metrics store.

**Implemented input contract (2026-09-01):** `POST /search/loop` generates
candidates via `strategy.RandomGenerator` and enqueues them one at a time
(`experiment.RunSearchLoop`). It accepts the intended OR-semantics controls:
- **max candidate count** — required (`MaxCandidates`, validated 2–200),
  always the backstop even if nothing else is set.
- **max wall-clock time** — optional (`MaxDurationSeconds`, validated
  60–3600 if set).
- **no-improvement-in-N** — optional (`NoImprovementLimit`). Currently tracked
  event-driven via a new `WorkerPool.AddObserver` (multiple observers, not
  just the one `SetObserver` hook the WebSocket hub already used) rather
  than polling the database before every candidate — the loop's observer
  gets pushed a notification the instant one of its own candidates
  completes, updates an in-memory best-score/steps-since-improvement pair,
  and the generator checks that before producing the next candidate. Because
  the production PostgreSQL queue has no bounded in-process buffer, generation
  can enqueue every requested candidate before a completion arrives. The
  observer is also removed as soon as generation ends. Therefore this input is
  not yet a reliable production stop condition; a bounded in-flight window or
  persisted search coordinator is required.
- **user-cancel** — still deferred. Nothing on the frontend calls a cancel
  endpoint today. Max-candidates still guarantees finite generation; the other
  two inputs must not be described as completed until their runtime semantics
  are corrected.

Within one loop run, generated candidates use a local dedup attempt (a
`seen` set keyed by sorted strategies + params + policy, in-memory, scoped
to that single run). After ten unsuccessful retries the current code accepts
the duplicate, so strict deduplication remains open.

`SEARCH_PROGRESS` (`{tested, total}`) is now real, not a placeholder:
`Result`/`BacktestJob` carry `SearchID`+`SearchTotal`, `Repository.List`
gained `ListBySearch`, and `Hub.JobUpdated` computes `tested` by counting
actual `COMPLETED`/`FAILED` rows for that search rather than hardcoding
`{0,1}`/`{1,1}`. This also unified the single-candidate `/search/start`
path onto the same mechanism (`SearchID = its own ID`, `SearchTotal = 1`)
instead of a separate, divergent code path.

This progress contract is complete only when the loop creates exactly
`SearchTotal` terminal rows. Early stop or enqueue failure can create fewer
rows while every row still advertises the requested maximum. There is currently
no search-run row or terminal event that revises the total or publishes
`STOPPED`; consumers can otherwise wait forever below 100%.

The frontend now calls `/search/loop`, sends all three controls and removes its
former timer-generated progress. It renders only WebSocket counts, treats HTTP
errors (including insufficient-candle 422) as `FAILED`, infers `COMPLETED` only
at `tested == total`, and is prepared for future additive
`{searchId,status,reason}` terminal fields.

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
- **Partial:** `SEARCH_PROGRESS` gives the frontend real terminal-candidate
  counts when every requested candidate is created. It does not yet fully
  answer loop health for early stops or distinguish concurrent searches.
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
- **Removed (2026-09-02):** a `RUNNING` result whose worker or process
  crashed mid-job used to be caught by a periodic sweep
  (`experiment.Sweeper` + `Repository.MarkStaleRunningFailed`, advisory-lock
  guarded so multiple instances wouldn't double-sweep). That's gone now —
  `PostgresQueue`'s lease/heartbeat/reclaim mechanism (ADR-0013) does the
  same job, faster (polls roughly every second vs. the sweep's one-minute
  tick) and more completely (reclaims and retries the job, not just marks
  it failed). Keeping both was redundant: the sweep's own SQL was still
  correct to run, but in practice it never won the race against the faster
  mechanism, so it was dead weight rather than genuine defense in depth.
  It only ever mattered as the *sole* safety net for `InMemoryQueue`, which
  isn't what's actually deployed.
