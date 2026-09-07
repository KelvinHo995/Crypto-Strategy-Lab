# 5. Strategy & Composition Flow

This is the module the spec calls "quan trọng nhất của hệ thống" (ch.6) and
where most of the extensibility grading happens (ch.41).

## Single strategy

```go
type Signal string
const ( Buy Signal = "BUY"; Sell Signal = "SELL"; Hold Signal = "HOLD" )

type Strategy interface {
    Name() string
    Analyze(candles []market.Candle) Signal
}
```
(`backend/internal/strategy/strategy.go`, matching `docs/contracts.md` §3.2.)

A strategy implementation (e.g. `MAStrategy`, `RSIStrategy`) only ever sees
candles in, a `Signal` out. It does not call Binance, does not touch the
database, does not know about charts or notifications — spec ch.7 states this
explicitly ("Không nên chứa: code gọi Binance / code lưu database / code vẽ
chart / code gửi notification"). This is what keeps a strategy testable in
isolation and swappable without touching anything downstream.

## Registration (plugin boundary)

```go
type Plugin struct {
    Name         string
    Default      Strategy
    Factory      StrategyFactory
    RandomParams RandomParamsGenerator
    Version      string
}
func (r *Registry) RegisterPlugin(plugin Plugin) error
```

Adding a new production strategy means: implement `Strategy` and its factory,
define its random parameter generator, then register one descriptor in
`cmd/server/main.go`. The descriptor also owns the static strategy
implementation version. Registration rejects duplicate names and incomplete
descriptors. Generic experiment code and the frontend do not need to change
— no `if strategy == "MA" { ... } else if strategy == "RSI" { ... }` dispatch
anywhere (spec ch.12, explicitly listed as the anti-pattern to avoid in
ch.44 "Hard-coded Strategy"). This is the concrete test spec ch.41 proposes:
"Hệ thống hiện có MA, RSI, Bollinger, SR. Hãy bổ sung MACD" should cost one
strategy-owned implementation unit plus one `RegisterPlugin` call in the
composition root.

`RandomGenerator` reads sorted descriptors and invokes each selected plugin's
`RandomParams` function. Backtester, Evaluator, WorkerPool, Ranking, and
RandomGenerator core therefore remain unchanged when another strategy is
registered. This is source-level startup registration, not runtime loading of
compiled plugins.

```
strategies/
    MAStrategy       ─┐
    RSIStrategy        │
    BollingerStrategy  ├─▶ RegisterPlugin(...) ─▶ Registry ─▶ (looked up by name, never by type-switch)
    SRStrategy         │
    SMCStrategy        ┘  (intentionally minimal — swing high/low structure break,
    MACDStrategy  ← new,     not full SMC theory; spec ch.11 only requires proving
    same shape, same          the plugin boundary holds, not trading-grade accuracy)
    registration call
```

The current server registers 7 single strategies: MA, RSI, Bollinger,
Support/Resistance, SMC, MACD and Sentiment. MACD is the concrete architecture
proof requested by the deck. The test-only `TestMomentum` descriptor proves
generation, factory construction, composite execution, deterministic seeded
generation, and registration of an eighth plugin without a count assumption.
`TestStrategyPluginRunsThroughExperimentPipeline` also declares a plugin
outside the strategy package and passes it through composition → backtest →
evaluation → ranking without changing those modules.
The frontend treats `GET /strategies` as authoritative: unknown names receive a
generic default-parameter form, so a new backend plugin appears in the picker
without adding hardcoded frontend metadata. Rich labels/parameter hints remain
an optional presentation enhancement, not a requirement for execution.

## Composition (composite strategies)

A `CandidateStrategy` is a named set of registered strategies plus a
combination policy:

```go
type CandidateStrategy struct {
    ID        string
    Instances []StrategyInstance // each has its own type, params and weight
    Policy    string             // "majority" | "weighted"
}
```

Two combination policies (spec ch.13–14):

- **Majority vote** — count BUY/SELL/HOLD across constituent signals, most
  votes wins.
- **Weighted score** — encode `BUY=+1, HOLD=0, SELL=-1`, compute
  `Σ(signal × weight)`, threshold the result (`score > 0.3 → BUY`,
  `score < -0.3 → SELL`, else `HOLD`).

Composition logic lives outside individual `Strategy` implementations — a
strategy never knows it's being combined with others. This is what lets the
composition policy itself be swapped (e.g. add a 3rd policy later) without
touching `MAStrategy` or `RSIStrategy`.

## What consumes a CandidateStrategy

`internal/experiment.WorkerPool` passes a `CandidateStrategy` to
`strategy.BuildFromCandidate`, which resolves each instance through the
`Registry` and returns one constructed strategy (a `CombinedStrategy` when
needed). The Backtester receives that resolved strategy, runs it against candle
windows, and produces trades. See
[06-search-backtest-flow.md](06-search-backtest-flow.md) for what happens
next (backtest → evaluate → rank).

See [ADR-0002](../adr/0002-strategy-plugin-registry.md) for why
interface+registry was chosen over alternatives (factory switch,
reflection-based discovery).
