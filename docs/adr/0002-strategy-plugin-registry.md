# ADR-0002: Strategies as a plugin (interface + registry), not a type switch

**Status:** Accepted
**Owner:** Strategy + Search (Person 2)

## Context

The spec's central extensibility test (ch.41) is literally "the system has
MA, RSI, Bollinger, SR today — add MACD" and grades how many components that
touches. Spec ch.44 names the anti-pattern directly:

```go
if strategy == MA { ... }
else if strategy == RSI { ... }
else if strategy == Bollinger { ... }
else if strategy == SR { ... }
```

This grows linearly with every new strategy and touches a shared file every
time, which is exactly the coupling the grading wants to see avoided.

## Decision

`Strategy` is an interface (`Name() string`, `Analyze([]Candle) Signal`).
Each strategy is its own type implementing it. A `Registry` holds strategies
by name. `RegisterPlugin(default, factory)` installs both lookup forms in one
validated operation; adding MACD is one composition-root call, with no edits to
Backtester, Evaluator, Leaderboard or frontend core.

```go
type Strategy interface {
    Name() string
    Analyze(candles []market.Candle) Signal
}
type Registry struct{ strategies map[string]Strategy }
func (r *Registry) RegisterPlugin(s Strategy, factory StrategyFactory) error
```

Every implementation also declares
`var _ Strategy = (*MAStrategy)(nil)` so a signature mismatch is a compile
error at the definition site, not a runtime surprise (PLAN.md §6 — this
check is explicitly "Có dùng").

## Alternatives considered

- **Factory function with a `switch` on a strategy-name string.** Rejected —
  this is the anti-pattern above, just moved into a `NewStrategy(name string)`
  function instead of inline; the coupling is identical.
- **Reflection-based auto-discovery** (scan a package for types implementing
  `Strategy`, register automatically). Rejected as unnecessary complexity —
  Go has no runtime package scanning without extra tooling, and explicit
  `Register()` calls are more debuggable (you can see exactly what's
  registered by reading `main.go`, not by inferring it from folder
  contents).
- **Config-driven strategies** (YAML describing indicator + thresholds,
  interpreted generically). Rejected for MVP — flexible, but adds a config
  interpretation layer the 2-week scope doesn't need; every MVP strategy is
  simple enough that a Go type is not meaningfully more work than a config
  schema, and a Go type gets compile-time safety the config approach
  wouldn't.

## Consequences

- **Positive:** the ch.41 scenario is implemented by `MACDStrategy`; production
  registration costs one new file + one `RegisterPlugin()` call. The
  cross-package architecture test drives an externally declared plugin through
  composition, backtest, evaluation and ranking.
- **Positive:** each strategy is independently unit-testable with a plain
  candle slice in, signal out — no mocking required.
- **Cost:** registration remains explicit in the composition root. A forgotten
  call means the plugin is absent, but duplicate names, nil implementations and
  missing factories now fail startup through `RegisterPlugin` rather than
  silently overwriting another plugin.
