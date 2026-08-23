# Architecture Decision Records

MADR-style: each file states Context → Decision → Alternatives considered →
Consequences. ADRs 0001-0006 and 0008-0011 map to the professor's own
suggested ADR list (`KienTrucDoAn_slide.pdf`, Appendix B) — see the
cross-reference table below. ADR-0007 (auth) is scope the team added on top,
not on the professor's list.

| ADR | Decision | Owner |
|---|---|---|
| [0001](0001-market-data-adapter.md) | Market data behind a Provider/Adapter boundary | Market Data |
| [0002](0002-strategy-plugin-registry.md) | Strategies as a plugin (interface + registry), not a type switch | Strategy + Search |
| [0003](0003-separate-backtester-evaluator.md) | Backtester and Evaluator are separate components | Experiment |
| [0004](0004-inprocess-job-queue-not-kafka.md) | In-process job queue behind a `Queue` interface, not Kafka/RabbitMQ | Strategy + Search / Experiment |
| [0005](0005-modular-monolith-not-microservices.md) | Modular monolith (one Go binary), not per-domain microservices | whole team |
| [0006](0006-separate-sentiment-service.md) | Sentiment analysis is a separate Python service, not a Go package | Market Data / Frontend |
| [0007](0007-simple-session-auth.md) | Minimal username/password accounts + short-lived (1h) JWT cookie, not a server-side session table | unassigned — new scope, see ADR |
| [0008](0008-websocket-for-realtime.md) | WebSocket for realtime UI, not polling/SSE | Market Data / Frontend |
| [0009](0009-experiment-provenance-storage.md) | How experiment/version provenance is stored | Experiment |
| [0010](0010-no-cqrs-event-sourcing.md) | CQRS and Event Sourcing are not used | whole team |
| [0011](0011-search-loop-stop-conditions.md) | Stop conditions and observability of the Search Loop | Experiment / Strategy + Search |
| [0012](0012-supabase-postgres-not-sqlite.md) | Supabase-hosted Postgres, not SQLite (shared team access, not performance) | unassigned — new scope, see ADR |

## Cross-reference to the professor's suggested ADR list

| Ours | Professor's list (Appendix B) |
|---|---|
| [0001](0001-market-data-adapter.md) | ADR-001 — Why MarketDataProvider + Adapter? |
| [0008](0008-websocket-for-realtime.md) | ADR-002 — Why WebSocket for realtime UI? |
| [0002](0002-strategy-plugin-registry.md) | ADR-003 — Why Strategy Plugin/Registry? |
| [0003](0003-separate-backtester-evaluator.md) | ADR-004 — Why separate Backtester and Evaluator? |
| [0004](0004-inprocess-job-queue-not-kafka.md) | ADR-005 — Why queue/worker (or why NOT)? |
| [0005](0005-modular-monolith-not-microservices.md) | ADR-006 — Why modular monolith vs microservices? |
| [0009](0009-experiment-provenance-storage.md) | ADR-007 — How experiment/version provenance is stored? |
| [0006](0006-separate-sentiment-service.md) | ADR-008 — Why separate News Collector and Sentiment Service? |
| [0010](0010-no-cqrs-event-sourcing.md) | ADR-009 — Why CQRS/Event Sourcing is used — or deliberately not used? |
| [0011](0011-search-loop-stop-conditions.md) | ADR-010 — Stop conditions and observability of Strategy Loop |

Status on every ADR here is **Accepted** unless stated otherwise — these are
decisions the team has already committed to for the 2-week build, not open
proposals. Ownership is the exception: auth (0007) is accepted scope, but no
team member is yet assigned to build it — see PLAN.md §0.
