# ADR-0011: Stop conditions and observability of the Search Loop

**Status:** Accepted
**Owner:** Experiment (Person 3), Strategy + Search (Person 2)

> **Implementation review (2026-09-02, updated same day):** terminal
> search-run state and atomic loop enqueue — both flagged open earlier
> today — are now shipped and live-verified against Supabase (see below).
> Job ID generation, also flagged, is fixed (UUID suffix, not just a
> timestamp). Still genuinely open: the production `PostgreSQL` queue has
> no backpressure, so no-improvement can overshoot by more than a few
> candidates; candidate dedup gives up after 10 retries and may still
> enqueue a duplicate. Both are documented below, not silently dropped.

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
  `searchId`, terminal `status` and `reason` are published today. Aggregate
  completed/failed counts and latency are exposed separately by `GET /metrics`;
  best score remains represented by `LEADERBOARD_UPDATE`.
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

The `tested == total` inference only resolves the loop's normal-completion
case (it generated its full `MaxCandidates`, so every eventual terminal row
counts toward the same total the frontend was told upfront). Early stop
(`MaxDuration`/`NoImprovementLimit`/cancellation/enqueue failure) enqueues
fewer than `MaxCandidates` jobs while every one of those jobs still carries
the original requested maximum as `SearchTotal` — `tested` would never reach
`total` on its own. `RunSearchLoop` closes this with an `onStopped(status,
reason string)` callback, fired exactly once, only on early stop (never on
normal completion, which the existing mechanism already covers): `status` is
`FAILED` if nothing was ever enqueued, `STOPPED` otherwise. The HTTP layer
wires this to `Hub.SearchLoopStopped`, which broadcasts an additive
`SEARCH_PROGRESS{searchId,status,reason}` the moment generation stops —
independent of whether already-enqueued jobs are still running. Live-verified
against Supabase: a `noImprovementLimit:1` run stopped generation at 3 of 10
requested candidates, broadcast `{status:STOPPED,reason:"no-improvement"}`
immediately, and `tested` kept climbing afterward (4/10) as the one
already-enqueued straggler finished — confirming "stopped generating" and
"finished running" are correctly signaled as separate things, per the OR-stop
rule below.

The frontend calls `/search/loop`, sends all three controls and has no
timer-generated progress. It renders only WebSocket counts, treats HTTP
errors (including insufficient-candle 422) as `FAILED`, infers `COMPLETED`
from `tested == total`, and consumes the additive `{searchId,status,reason}`
terminal fields the backend now publishes for `STOPPED`/`FAILED`.

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
- **Positive:** `SEARCH_PROGRESS` gives the frontend real terminal-candidate
  counts on normal completion (`tested == total`) and an explicit
  `{status,reason}` terminal signal on early stop — both paths now
  distinguishable, live-verified against Supabase.
- **Positive:** `/search/loop` now writes its `PENDING` row and job through
  `PostgresQueue.EnqueuePending` in one transaction, same as `/search/start`
  — closes the crash window that existed when it wrote the two as separate
  calls.
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
