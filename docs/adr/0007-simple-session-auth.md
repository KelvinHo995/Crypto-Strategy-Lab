# ADR-0007: Minimal username/password accounts + short-lived JWT (1h)

**Status:** Accepted
**Owner:** unassigned — new scope, not part of the original 14-day plan (see [PLAN.md](../../PLAN.md) §0)

## Context

The system was originally scoped as single-tenant, no auth (see
[01-context.md](../architecture/01-context.md)'s original non-goals). That
assumption breaks once the app is reachable beyond the team — a demo link
shared for grading, or hosted anywhere public. Every `POST /search/start`
call is expensive: it enqueues real work onto the Job Queue/worker pool
(backtesting N candidates against historical data,
[06-search-backtest-flow.md](../architecture/06-search-backtest-flow.md)).
An anonymous visitor triggering repeated searches is a real cost, not a
theoretical one.

The team's explicit direction: real user accounts (not just a shared
secret), but "very very simple" — this is a 2-week build, not a product with
a security team. On revocation specifically: the team decided not to build
any admin/manual revoke tooling — logout plus a short expiry is enough, and
a 1-hour expiry was judged an acceptable exposure window on its own.

## Decision

- `users` table: `id`, `username` (unique), `password_hash` (bcrypt),
  `created_at`. No email, no profile fields, no roles.
- Session token is a **signed JWT (HS256, server-side secret via env var)**,
  not a server-side session record:

  ```
  claims: { sub: userID, iat, exp }   // exp = iat + 1h
  ```

- Delivered as an **httpOnly, SameSite=Lax cookie** — the browser sends it
  automatically on every subsequent request, including the WebSocket
  upgrade handshake (still an HTTP request at connect time), so the
  frontend never manually attaches a token to anything.
- Middleware validates signature + `exp` only — **no database lookup** per
  authenticated request, no session table to maintain.
- **One middleware, applied uniformly**: every route requires a valid,
  unexpired token **except** `/health`, `POST /auth/register`, and
  `POST /auth/login`. No curated public/private route list.
- `POST /auth/logout` is a client-side action (server responds with a
  cookie-clearing `Set-Cookie`) — there's no server-side row to delete, so a
  token already copied out of the browser stays technically valid until it
  naturally expires (≤1h) even after "logout." Accepted tradeoff, not an
  oversight — see Consequences.
- Registration is self-service (`username` + `password`, bcrypt-hashed) —
  the goal is stopping anonymous/unauthenticated access, not restricting who
  can have an account.

## Alternatives considered

- **Opaque server-side session token + `sessions` table**, revoked by
  deleting a row. Rejected: it buys early revocation (kill a session before
  its expiry), which the team already decided isn't worth building tooling
  around — a 1-hour cap already bounds exposure to roughly the same window
  in practice, so the extra table and per-request DB lookup aren't earning
  their cost here.
- **Shared secret / HTTP Basic Auth**, no per-user identity. Rejected — a
  shared secret can't answer "which user ran this experiment," and
  per-user attribution is a natural extension of the provenance the
  Experiment domain already tracks (`StrategyVersions`, spec ch.36) — "who
  ran it" belongs next to "what it ran."
- **OAuth / third-party login** (Google, GitHub, etc.). Rejected — no driver
  requires federated identity for a 4-person team's own demo; adds a
  third-party dependency and callback flow for zero benefit at this scope.
- **Plaintext or unsalted-hash passwords**, to save a dependency. Rejected —
  this is the one place "simple" must not mean "insecure." bcrypt is a
  single library call, not meaningfully more complex than a weaker
  alternative, and storing recoverable passwords is a real vulnerability
  even in a class project.

## Consequences

- **Positive:** no `sessions` table, no DB hit to validate a request — one
  less table, one less piece of state to keep consistent. Matches "very
  very simple" better than the session-table alternative once early
  revocation is off the table as a requirement.
- **Positive:** stateless validation is trivially safe if the process
  restarts — not a current driver, but free.
- **Cost, accepted:** logout doesn't immediately invalidate a token
  server-side; worst case is a ≤1h exposure window. There is no
  admin/manual-revoke path — if a token ever needs to die faster than 1h
  during a demo, the only lever is rotating the signing secret, which
  invalidates *all* outstanding tokens, not just one.
- **New requirement:** a JWT signing secret must be generated and kept in
  env config, not committed to the repo.
- **Cost:** new package needed (`internal/auth` or similar) — see
  [02-modules.md](../architecture/02-modules.md) — not yet assigned to
  anyone in the current 14-day plan (buffer days 8–9 are the natural slot,
  PLAN.md §2).
- **Cost:** frontend needs a login/register form and must handle `401`
  (redirect to login) globally.
- **Consequence for the demo (spec ch.46):** the scripted demo flow now
  starts with a login step before "Bước 1 – Mở BTCUSDT" — update the demo
  script accordingly when rehearsing.
