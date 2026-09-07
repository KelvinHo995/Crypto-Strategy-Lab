# ADR-0009: How experiment/version provenance is stored

**Status:** Accepted
**Owner:** Experiment (Person 3)

## Context

Spec ch.36 requires every leaderboard result to be traceable back to the
exact strategy code, parameters, dataset, and (if applicable) sentiment
model version that produced it: "Experiment #122 luôn biết chính xác nó đã
sử dụng strategy nào." The professor's deck frames this as **Reproducibility**
— a named rubric criterion ("Top-K có link về exact experiment
config/version?") — and uses an MLOps "production batch label" analogy: if
a batch of predictions is bad, you recall that exact batch, not the whole
product line. Without stored provenance, a leaderboard entry becomes
unverifiable the moment any constituent strategy's code changes after the
fact.

## Decision

Every `Result` row (the write side of the Experiment domain) stores aggregate
provenance inline at write time:

```go
type Result struct {
    ID               string
    CandidateID      string                   // exact CandidateStrategy spec used
    Pair             string                   // market symbol, e.g. BTCUSDT
    Timeframe        string                   // candle interval actually used
    Instances        []StrategyInstance       // strategy types, params, weights
    Policy           string                   // majority or weighted
    StrategyVersions map[string]string        // strategy implementation versions
    SentimentModels  []SentimentModelIdentity // model outputs actually consumed
    DatasetPeriod    string                   // exact historical window backtested
    Return, MDD      float64
    TradeCount       int
    Status           string
    CreatedAt        int64
}
```

`Pair`, `Timeframe`, and `DatasetPeriod` form the market dataset identity.
`DatasetPeriod` is only the historical `from-to` range and therefore cannot
identify a symbol or interval by itself. Legacy rows keep empty pair/timeframe
values so missing provenance is visible instead of being replaced by a guess.
Migration `0014` recovers coordinates only from retained durable job payloads
or persisted trade rows; an unknowable legacy timeframe remains empty.

`StrategyVersions["Sentiment"]` identifies the Go strategy implementation;
it is not used as a proxy for the Python model. During a backtest, each
runtime-built `SentimentStrategy` records the model name/version attached to
successful persisted observation lookups. The worker stores the unique list
in deterministic name/version order. Multiple identities are retained when a
historical range spans a model change; an empty list means no sentiment model
output was consumed (for example, fallback after missing/stale data).

This is plain CRUD — provenance is captured once, at the moment the result
is written, not reconstructed later by correlating logs or replaying events
(see [ADR-0010](0010-no-cqrs-event-sourcing.md) for why event replay isn't
used for this).

Trade rows are also part of the reproducibility record. They are written before
the aggregate result becomes `COMPLETED`; a failed trade write produces a
`FAILED` result rather than a successful but irreproducible leaderboard row.

## Alternatives considered

- **Reconstruct provenance after the fact**, by correlating a separate
  strategy-version-history log against the result's timestamp. Rejected —
  fragile: breaks the moment strategy code is edited without a version bump
  logged at exactly that moment, and requires a join/correlation step every
  time provenance needs to be checked instead of it being immediately
  available on the row.
- **Store only strategy names, not versions** (assume "MA" always behaves
  identically). Rejected — directly violates spec ch.36; a strategy's
  internal logic can change between two experiment runs even though its
  registered name doesn't.
- **Event Sourcing** (derive provenance by replaying
  `ExperimentCreated → CandidateAssigned → BacktestStarted → ... →
  LeaderboardPromoted`). Considered and rejected — see
  [ADR-0010](0010-no-cqrs-event-sourcing.md); would give a full audit trail
  of every state transition, not just the final provenance snapshot, but
  that's more than spec ch.36 actually asks for, at a real cost in schema
  evolution and replay complexity.

## Consequences

- **Positive:** any leaderboard row is self-describing — answering "what
  exactly produced this?" needs zero joins or replay, just reading the row.
- **Positive:** directly satisfies spec ch.36 and the deck's Reproducibility
  rubric question with a one-line query.
- **Positive:** model provenance comes from the stored sentiment observations,
  not a version guessed before the backtest starts.
- **Cost / open risk:** every strategy needs an explicit version identifier
  that a developer manually bumps when its logic changes — there is
  currently no automatic detection or enforcement that a strategy's version
  was bumped when its code changed. A forgotten version bump silently
  produces a mislabeled (not missing) provenance record, which is worse
  than an obviously-missing one. Worth a lightweight guard later (e.g., a
  test that fails if a strategy's logic hash changes without its declared
  version changing) — not built for MVP.
