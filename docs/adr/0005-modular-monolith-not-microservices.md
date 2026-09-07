# ADR-0005: Modular monolith (one Go binary), not per-domain microservices

**Status:** Accepted
**Owner:** whole team

## Context

The system has four clear domains — Market Data, Strategy/Search, Experiment
(backtest/evaluate/rank), and (arguably) News/Sentiment. Each domain has one
owner (see root README). The question is whether each domain should be its
own deployed service (microservices) or separate packages within one process
(modular monolith), for a 2-week, 4-person build with one demo environment.

## Decision

Market Data, News, Go-side Sentiment, Strategy, and Experiment live as focused
packages (`internal/market`, `internal/news`, `internal/sentiment`,
`internal/strategy`, `internal/experiment`) inside the Go backend, composed in
`cmd/server` (with `cmd/backfill` and `cmd/news-ingest` as manual utilities).
Only Python sentiment inference is split into its own deployed service — see
[ADR-0006](0006-separate-sentiment-service.md) for why that one domain is
different.

Package boundaries are enforced the same way process boundaries would be:
one owner per package, no reaching into another package's internals, defined
interfaces between them (`Strategy`, `Candle`) — but calls between them are
Go function calls, not network requests.

## Alternatives considered

- **One microservice per domain** (Market Data service, Strategy service,
  Experiment service, each with its own HTTP API). Rejected for this
  scope — PLAN.md §6 lays out the supporting reasoning explicitly: no
  framework beyond `net/http` (~6 endpoints doesn't justify one), no
  Kubernetes (no orchestration need for 1 backend + 1 frontend + 1 sentiment
  service), no Redis (single instance, no shared-state-across-replicas
  need). Splitting Market/Strategy/Experiment into separate services would
  multiply operational surface (3 more processes to run, deploy, and debug
  during a demo) without a real scaling or team-autonomy need forcing it —
  a 4-person team sharing one repo doesn't need service-level deploy
  independence the way separate teams would.
- **Everything in one undifferentiated package** (no focused `internal/*`
  domain packages). Rejected —
  this is the God Service anti-pattern the spec calls out (ch.44); package
  boundaries are what make ownership, testability, and the "add MACD"
  extensibility test (ch.41) meaningful even without process boundaries.

## Consequences

- **Positive:** fast local dev loop (one process to run), no cross-service
  network calls to mock or reason about between Market/Strategy/Experiment,
  simpler demo (fewer moving pieces to fail on stage).
- **Positive:** package-level `internal/` visibility gives real enforcement
  against cross-domain leakage — Go's compiler rejects an import that
  reaches into another package's unexported internals.
- **Cost:** package boundaries are weaker than process boundaries — nothing
  stops a determined author from importing across domains in ways an
  interface should have prevented; this is enforced by code review
  ownership (root README: "cross-domain changes need that person's
  review"), not by the runtime.
- **Explicit non-goal:** independent scaling of, say, Experiment workers
  separately from the HTTP server. If backtest load ever needs to scale
  independently of request-serving load, that's the point at which
  splitting `internal/experiment` into its own service becomes justified —
  not before.
