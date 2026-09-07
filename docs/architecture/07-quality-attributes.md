# 7. Quality Attributes & Architectural Drivers

Spec ch.32 frames these as the actual grading criteria — "architectural
drivers", not feature checkboxes. Each one below states the pressure, the
decision made in response, and where that decision is documented in depth.

## Modifiability

**Pressure:** add `MACDStrategy` without touching 20 other modules
(spec ch.32.1).
**Decision:** `Strategy` interface + atomic `Registry.RegisterPlugin()`
([05-strategy-flow.md](05-strategy-flow.md), [ADR-0002](../adr/0002-strategy-plugin-registry.md)).
**Test:** spec ch.41 scenario — count how many generic consumers a new strategy
touches. The strategy owns its implementation, factory, random parameter
generator, and version descriptor; the composition root adds one registration.
MACD is the real example, while test-only `TestMomentum` proves generation,
construction, composite execution, seeded determinism, and an eighth plugin
without modifying generic consumers.

## Scalability

**Pressure:** 10 strategies today → 100,000 candidate strategies later
(spec ch.32.2, ch.43).
**Decision:** Job Queue + Worker pool, worker count is the scaling knob, not
a rewrite ([06-search-backtest-flow.md](06-search-backtest-flow.md),
[ADR-0013](../adr/0013-postgres-durable-queue-and-local-caches.md)).
**Implemented:** PostgreSQL workers claim safely across backend instances with
`SKIP LOCKED`; durable leases survive process restarts. A dedicated broker is
still unnecessary at current throughput. The current API caps a run at 200
candidates. No-improvement runs bound unresolved work by their configured
limit, but runs without that condition have no universal in-flight cap; a
100,000-candidate run is not supported today.

## Realtime

**Pressure:** Binance market data → indicator → strategy signal → UI with
low latency, without the frontend polling (spec ch.32.3).
**Decision:** WebSocket push, one connection, tagged message types
(`CANDLE_UPDATE`, `SEARCH_PROGRESS`, `LEADERBOARD_UPDATE`) —
[04-realtime-flow.md](04-realtime-flow.md).

## Reliability

**Pressure:** Binance connection drops — does the system stay usable
(spec ch.32.4, ch.40 Q7)?
**Decision:** reconnect/retry logic owned entirely inside
`internal/market`, isolated from strategy/experiment/frontend. Local tests
prove disconnect, bounded backoff, reconnect, resumed event delivery, and
clean cancellation. Upstream Binance reconnect state is not yet exposed to the
frontend, so visible upstream-staleness reporting remains a limitation.
**Answered:** a job whose on-demand candle fetch comes back short
(`< experiment.MinCandlesForBacktest`) is retried through the same
`Queue.Nack`/backoff/lease-exhaustion path as every other worker failure,
not failed outright — see `internal/experiment/worker.go`'s `run()`. This
covers the "mid-backfill" case (transient, a retry a few seconds later
sees the rest of the data) without needing to special-case it: a request
against a range with genuinely no data just exhausts retries and lands on
`FAILED` the same way, a few seconds later instead of immediately.

## Performance

**Pressure:** 1,000 strategies to backtest — sequential loop vs. concurrent
workers (spec ch.32.5).
**Decision:** same as Scalability above — Job Queue + Workers.
**Measurement:** `BACKTEST_WORKERS` is configurable from 1–64;
`go run ./cmd/perf -candidates 1000 -candles 2000 -workers 1,3 -repeat 3` runs
the same in-memory historical workload at both concurrency levels, discards a
warm-up, and prints per-run plus median duration/jobs per second/speedup.
Production `GET /metrics` complements that controlled proof
with live queue wait, execution time, throughput and queue state.
The recorded three-run baseline and its limits are in
[`performance-evidence.md`](../performance-evidence.md).

## Replaceability (spec ch.32.6 calls this "Maintainability" — same test)

**Pressure:** can an implementation be swapped without its consumers
changing? Two concrete instances:
1. Swap `Random Search` for `Genetic Search` without the Backtester
   changing (spec ch.32.6, spec ch.42's scenario).
2. Swap `PostgresQueue` for Redis Streams without the
   worker pool, `StrategyGenerator`, or `Backtester` changing, if
   distributed workers ever become a real requirement.
**Decision:** both are interfaces consumers depend on, never the concrete
implementation — `StrategyGenerator` (`generate() -> CandidateStrategy`) and
`Queue` (`Enqueue`/`Dequeue`/`Ack`/`Nack` over `BacktestJob`) —
[06-search-backtest-flow.md](06-search-backtest-flow.md),
[ADR-0013](../adr/0013-postgres-durable-queue-and-local-caches.md).
**Test:** for either swap, downstream consumers remain unchanged. Source-level
composition still selects the implementation in the server wiring; this is
dependency injection, not runtime implementation discovery.

## Observability

**Pressure:** is the loop running? how many candidates tried? how many
job failures? who's #1 right now? (spec ch.32.7)
**Status:** implemented for single and continuous searches: workers persist
`PENDING → RUNNING → COMPLETED/FAILED`; the WebSocket hub broadcasts
`SEARCH_PROGRESS` and a ranked Top-10 `LEADERBOARD_UPDATE`. Authenticated
`GET /metrics` exposes queue depth/running/failed rows, process-lifetime
completed/failed counters, throughput, average and p50/p95 queue wait/backtest
execution time. `GET /ready` checks PostgreSQL separately from liveness, and
each HTTP response carries an `X-Request-ID` that is included in a duration log.
`GET /metrics/prometheus` exports the worker/queue/latency signals using the
Prometheus exposition format, and normal worker transitions are logged with
`search_id`, `job_id`, `candidate_id` and status. Deploying a Prometheus server,
distributed tracing and alert routing remain operational post-MVP work.

## Reproducibility

**Pressure:** can a leaderboard result be traced back to the exact strategy
code, parameters, dataset, and model version that produced it? (deck
rubric: "Top-K có link về exact experiment config/version?", spec ch.36).
**Decision:** `Result` rows store `Pair`, `Timeframe`, `DatasetPeriod`,
`CandidateID`, strategy instances/parameters, `StrategyVersions`, and runtime
`SentimentModels`; generated trades are linked through `experiment_trades` —
[ADR-0009](../adr/0009-experiment-provenance-storage.md), provenance
section in [06-search-backtest-flow.md](06-search-backtest-flow.md).
**Explicitly not done:** Event Sourcing / a full state-transition audit
trail — [ADR-0010](../adr/0010-no-cqrs-event-sourcing.md) explains why
state-only storage already answers what's actually asked.
**Open risk:** version bumps on strategy code changes are a manual
discipline, not enforced automatically — see ADR-0009's Consequences.

## Security / Access Control

**Pressure:** an anonymous visitor must not be able to trigger a backtest —
every search run enqueues real work onto the Job Queue/worker pool
([06-search-backtest-flow.md](06-search-backtest-flow.md)), and the app may
be reachable beyond the team (grading, a shared demo link).
**Decision:** minimal username/password accounts + a short-lived (1h) signed
JWT cookie, gating the whole API by default (one middleware, not a curated
public/private route list) — [ADR-0007](../adr/0007-simple-session-auth.md).
**Explicitly not done:** roles/permissions, password reset, OAuth/SSO — no
driver requires them at this scope.
**Implemented:** `internal/auth` owns bcrypt/JWT/user persistence and
`internal/httpx` applies one cookie-verification middleware to every
non-public route. `JWT_SECRET` is required at server startup.

## Deferred extensions (explicitly out of scope for MVP)

Two ideas surfaced (from the professor's supplementary notes) that are
deliberately **not** being built in the 2-week window, with reasoning kept
here rather than as ADRs — an ADR justifies a decision that was made and
acted on; these are decisions to *not* act, which belong with the other
scope cuts in [PLAN.md](../../PLAN.md)'s "Ngoài phạm vi 2 tuần" list.

- **Natural language / website link → auto-generate strategy.** A user
  types a description or pastes a link, and the system synthesizes a
  `CandidateStrategy` via an LLM. Cut because: (1) it's a large surface —
  LLM prompt/output validation, ensuring generated output actually conforms
  to the `Strategy` interface, plus its own UI — for a 4-person/2-week
  build; (2) it doesn't prove anything about the architecture that
  `StrategyGenerator` (Random → Domain-guided,
  [06-search-backtest-flow.md](06-search-backtest-flow.md)) doesn't already
  demonstrate — the interface is what's graded, not how creative the
  generator is. If time allows post-MVP, it plugs in as one more
  `StrategyGenerator` implementation, no architecture change required.
- **LLM-based HTML tag extraction with cached schema for the News
  Crawler.** Cut because a fixed parser per structured source (RSS feed,
  News API) already satisfies the `NewsProvider` abstraction
  ([ADR-0006](../adr/0006-separate-sentiment-service.md)'s sibling
  decision, spec ch.28) without needing raw-HTML scraping at all for MVP.
  Worth revisiting only if a required news source has no structured feed.

## Documentation

**Pressure:** deck rubric asks "Context/Container/Dynamic view nhất quán
với code?" — is this documentation an accurate map of the system, or has
it drifted from what's actually built?
**Decision:** the `docs/architecture/` + `docs/adr/` split (trimmed arc42
sections + one-file-per-decision ADRs), cross-referenced by chapter/slide
number against both the project spec (`Crypto Strategy Lab – Đồ án cuối
kỳ.pdf`) and the professor's teaching deck (`KienTrucDoAn_slide.pdf`), so
each claim traces to a source instead of being asserted from memory.
**Current practice:** these pages and ADR implementation notes are checked
against current source, migrations, and tests, with historical decisions marked
as superseded or evolved where appropriate. Documentation drift remains a
maintenance risk whenever behavior changes without the corresponding doc edit.

## Trade-off reasoning

**Pressure:** deck rubric: "mỗi công nghệ có driver và consequence?" — every
non-trivial choice should state what it costs, not just what it buys
("không có free lunch", deck slide 69).
**Where this lives:** every ADR's Consequences section states both the
benefit and the cost of that decision — not duplicated here. See in
particular [ADR-0013](../adr/0013-postgres-durable-queue-and-local-caches.md)
(durable queue/local caches vs. polling and per-process state),
[ADR-0005](../adr/0005-modular-monolith-not-microservices.md) (monolith:
simplicity vs. no independent scaling), and
[ADR-0010](../adr/0010-no-cqrs-event-sourcing.md) (CRUD: simplicity vs. no
audit trail).
**Test:** pick any ADR at random during vấn đáp and be able to state its
cost, not just its benefit, without re-reading it first.

## Anti-patterns this architecture is designed to avoid (spec ch.44)

| Anti-pattern | Where it would show up | Why the current package boundaries prevent it |
|---|---|---|
| God Service | one `TradingService` doing Binance calls + RSI math + crawling + ML + backtest + ranking + DB + WS | `internal/market` / `internal/strategy` / `internal/experiment` are separate packages with one owner each |
| Hard-coded strategy dispatch | `if strategy == MA ... else if ...` | `Registry.RegisterPlugin()` — see [05-strategy-flow.md](05-strategy-flow.md) |
| Business logic in frontend | React becoming the authoritative backtest/ranking engine | Backend persists and ranks authoritative results; the frontend only mirrors the score formula for display/sorting |
| Strategy → Database directly | `RSIStrategy` calling Postgres directly | `Strategy.Analyze(candles) Signal` — no I/O in the interface at all |
| Crawler tightly coupled to ML model | `Crawler → BERT model` inline | `sentiment-service` is a separate deployable behind a REST boundary — [ADR-0006](../adr/0006-separate-sentiment-service.md) |

## The 8 central questions (spec ch.40) — where each is answered

1. Strategy mới thêm vào hệ thống thế nào? → [05-strategy-flow.md](05-strategy-flow.md)
2. Search algorithm mới thêm thế nào? → [06-search-backtest-flow.md](06-search-backtest-flow.md)
3. Market Data Provider mới thêm thế nào, có sửa frontend không? → [04-realtime-flow.md](04-realtime-flow.md), [ADR-0001](../adr/0001-market-data-adapter.md)
4. Backtest 100 → 100,000 thì kiến trúc thay đổi thế nào? → Scalability section above
5. News Service lỗi thì Chart còn chạy không? → [ADR-0006](../adr/0006-separate-sentiment-service.md) (must degrade gracefully, not cascade)
6. Sentiment Model thay đổi thì Strategy Engine bị ảnh hưởng không? → separate `SentimentModels` runtime provenance in [06-search-backtest-flow.md](06-search-backtest-flow.md)
7. Binance WebSocket disconnect thì phục hồi thế nào? → Reliability section above
8. Kết quả trên Leaderboard truy được version strategy nào? → Provenance section in [06-search-backtest-flow.md](06-search-backtest-flow.md)

Each team member should be able to answer the question tied to their owned
domain (see root README ownership table) without reading from this doc —
this file is the fallback reference, not the primary source of
understanding.
