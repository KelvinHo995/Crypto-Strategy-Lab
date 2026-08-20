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
type Registry struct { strategies map[string]Strategy }
func (r *Registry) Register(s Strategy)
```

Adding a new strategy means: implement `Strategy`, call `Register`. Nothing
in `internal/experiment`, `cmd/server`, or the frontend should need to change
— no `if strategy == "MA" { ... } else if strategy == "RSI" { ... }` dispatch
anywhere (spec ch.12, explicitly listed as the anti-pattern to avoid in
ch.44 "Hard-coded Strategy"). This is the concrete test spec ch.41 proposes:
"Hệ thống hiện có MA, RSI, Bollinger, SR. Hãy bổ sung MACD" should cost one
new file + one `Register` call.

```
strategies/
    MAStrategy       ─┐
    RSIStrategy        │
    BollingerStrategy  ├─▶ register(...) ─▶ Registry ─▶ (looked up by name, never by type-switch)
    SRStrategy         │
    SMCStrategy        ┘  (intentionally minimal — swing high/low structure break,
    MACDStrategy  ← new,     not full SMC theory; spec ch.11 only requires proving
    same shape, same          the plugin boundary holds, not trading-grade accuracy)
    registration call
```

MVP ships 5 single strategies: MA, RSI, Bollinger, Support/Resistance, and
SMC. SMC is deliberately scoped down — same registration mechanism as the
other four, just a simpler rule set — so it counts as a 5th registered
strategy without becoming its own time sink.

## Composition (composite strategies)

A `CandidateStrategy` is a named set of registered strategies plus a
combination policy:

```go
type CandidateStrategy struct {
    ID         string
    Strategies []string          // e.g. ["MA20", "RSI14", "SupportResistance"]
    Params     map[string]any
    Policy     string            // "majority" | "weighted"
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

`internal/experiment`'s Backtester takes a `CandidateStrategy`, resolves each
named strategy via the `Registry`, runs them against a candle window, applies
the policy, and produces trade signals. See
[06-search-backtest-flow.md](06-search-backtest-flow.md) for what happens
next (backtest → evaluate → rank).

See [ADR-0002](../adr/0002-strategy-plugin-registry.md) for why
interface+registry was chosen over alternatives (factory switch,
reflection-based discovery).
