# ADR-0002: Strategies as source-level plugins

**Status:** Accepted
**Owner:** Strategy + Search (Person 2)

## Context

The central extensibility scenario is adding a strategy such as MACD without
adding strategy-name branches to search, backtesting, evaluation, workers, or
ranking. A registry that stores only implementations and factories is
insufficient when random parameter generation and version metadata remain in
separate central switches or maps.

## Decision

Each strategy implements the small `Strategy` interface. Startup registers one
complete descriptor:

```go
type Plugin struct {
    Name         string
    Default      Strategy
    Factory      StrategyFactory
    RandomParams RandomParamsGenerator
    Version      string
}
```

`Registry.RegisterPlugin(Plugin)` validates and atomically registers the
implementation, factory, parameter generator, and implementation version.
`RandomGenerator` discovers sorted descriptors through `Registry.Plugins()`;
it never switches on concrete strategy names. Experiment provenance resolves
strategy versions through `Registry.VersionsFor()` from the same descriptor.
Sentiment model identity remains separate runtime provenance.

Adding a strategy now requires:

1. Implement `Strategy`.
2. Implement its factory.
3. Define a valid `RandomParamsGenerator` (or explicitly use
   `NoRandomParams`).
4. Register one `Plugin` descriptor in `cmd/server/main.go`.

No change is required in `RandomGenerator`, Backtester, Evaluator, WorkerPool,
or Ranking. Registration is source-level and occurs at startup; this decision
does not provide runtime loading of compiled plugins.

## Alternatives considered

- **Strategy-name switch:** rejected because every strategy modifies generic
  search code and violates the Open/Closed Principle.
- **Reflection-based discovery:** rejected because Go does not provide useful
  package scanning here and explicit startup composition is easier to audit.
- **Configuration-interpreted strategies:** rejected for the MVP because it
  adds a language and validation layer without improving the required proof.

## Consequences

- A plugin owns its construction, random search space, and static
  implementation version.
- Sorted descriptor discovery plus a seeded random source makes candidate
  generation reproducible for a fixed seed and plugin set.
- Duplicate names and incomplete descriptors fail during registration.
- Plugin authors must keep generated parameters within factory-valid ranges.
- The composition root still changes by one registration line for each plugin.
