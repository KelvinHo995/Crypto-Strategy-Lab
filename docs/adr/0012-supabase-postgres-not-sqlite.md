# ADR-0012: Supabase-hosted Postgres, not SQLite

**Status:** Accepted
**Owner:** unassigned — see PLAN.md §0

## Context

SQLite was the original choice (PLAN.md §0, §3c) — an embedded file needs no
separate service to run, which fit the team's general "no unnecessary
operational surface" philosophy (§6). That reasoning holds for a
single-developer or single-instance setup, but breaks down for this team
specifically: **4 people need to see the same leaderboard, the same
`experiments` table, the same backfilled `candles`.** A SQLite file is local
to whichever machine runs the backend — it doesn't solve "everyone sees the
same data" at all, regardless of how the file itself performs.

This is explicitly *not* a performance decision. The actual write load on
this database — roughly one `Result` row per completed backtest, every ~2s
per worker (spec ch.43) — is nowhere near where SQLite's single-writer
serialization would matter; that reasoning was considered and rejected in
an earlier discussion. The driver here is team data-sharing, full stop.

## Decision

Use a Supabase-hosted Postgres instance, reached the same way any Postgres
database would be — a connection string (env var, not committed), a
standard Go Postgres driver, through the same `Repository` interface
already used to keep persistence swappable (PLAN.md §6: "DI qua interface —
Có dùng"). This is a drop-in implementation swap behind an existing seam,
the same shape as the `Queue` interface in
[ADR-0004](0004-inprocess-job-queue-not-kafka.md) — no handler, strategy, or
experiment code changes because of this.

**Scope of what "using Supabase" means here, explicitly:**
- Supabase **only as a Postgres host.** Plain SQL over a connection string.
- **Not** Supabase Realtime — the team already designed and owns its own
  WebSocket/event layer ([ADR-0008](0008-websocket-for-realtime.md)), which
  is itself graded architecture (spec ch.32.3); outsourcing it would
  undercut that.
- **Not** Supabase Auth for now — [ADR-0007](0007-simple-session-auth.md)'s
  hand-rolled JWT decision stands; this ADR doesn't reopen it. (If that
  changes later, it's a separate decision, not implied by this one.)
- The frontend still never talks to Supabase directly — only the Go backend
  does, preserving the existing "frontend only talks to backend" rule
  ([02-modules.md](../architecture/02-modules.md)).

## Alternatives considered

- **Keep SQLite.** Rejected — doesn't solve the actual problem (shared
  team access); the file lives on one machine.
- **Self-hosted Postgres** (a shared server the team runs itself, e.g. via
  docker-compose on a VPS). Rejected — trades a database-hosting problem
  for a server-provisioning-and-securing problem; more ops surface for a
  4-person/2-week team than a managed instance, with no benefit over it.
- **Other managed Postgres providers** (Neon, Railway, RDS, etc.). Not
  rejected on technical grounds — any of them would satisfy the same
  driver. Supabase was picked because the team was already evaluating it;
  this decision is about "managed Postgres, reached only by the backend"
  more than it's specifically about Supabase's other features.
- **Supabase Realtime and Supabase Auth as full replacements for the
  existing WebSocket layer and hand-rolled JWT.** Considered separately,
  not adopted here — see Decision's scope note above.

## Consequences

- **Positive:** one shared source of truth — every team member's local
  backend instance points at the same data, no more "works on my machine"
  from divergent local SQLite files.
- **Positive:** zero database ops for the team — no server to provision,
  patch, or back up themselves.
- **Cost:** local dev now requires network access — no fully offline dev
  loop, and a Supabase outage blocks everyone's backend, not just one
  person's. Not a concern for a 2-week class project, but a real dependency
  now that didn't exist with a local file.
- **Cost — schema migration detail:** SQLite's `INTEGER` affinity
  transparently stores 64-bit values; Postgres's `INTEGER` type is 32-bit
  (max ~2.1 billion) and **overflows on unix-millisecond timestamps**
  (current unix ms is already ~1.7 trillion). Every millisecond-timestamp
  column (`candles.open_time`, `experiments.created_at`,
  `users.created_at`) must be declared `BIGINT`, not `INTEGER`, in the
  Postgres schema — PLAN.md §3c has been updated accordingly. This is the
  kind of silent-truncation bug that's easy to miss until it's in
  production.
- **Cost:** the connection string is a shared secret across 4 people now —
  needs a safe distribution method (each person's own `.env.local`,
  gitignored; not pasted into chat or committed) rather than a single
  hardcoded value anyone could accidentally commit.
