# Performance architecture evidence

This is a repeatable architecture proof for the rubric's “1,000 backtests:
sequential or parallel?” question. It measures the real `WorkerPool` → strategy
composition → `Backtester` → `Evaluator` path with deterministic in-memory
candles/repository so database and Binance latency do not hide worker scaling.
Production behavior is measured separately by authenticated `GET /metrics`.

## Baseline run — 2026-09-06

Environment: Windows/amd64, Go 1.27.0, 12 logical processors.

```powershell
go run ./cmd/perf -candidates 1000 -candles 2000 -workers 1,3 -repeat 3
```

| Run | 1 worker | 3 workers |
|---|---:|---:|
| 1 | 6,556.53 jobs/s | 12,168.54 jobs/s |
| 2 | 6,641.65 jobs/s | 14,290.76 jobs/s |
| 3 | 6,101.84 jobs/s | 15,286.03 jobs/s |
| Median | **6,556.53 jobs/s** | **14,290.76 jobs/s** |

The median throughput improved by approximately **2.18×** when changing only
the worker count from 1 to 3. This is not a production-capacity promise: the
controlled proof intentionally excludes PostgreSQL/network contention and uses
one machine. It demonstrates that the concurrency knob works without changing
Backtester, Evaluator, strategy implementations or queue contracts.

For a production run, set `BACKTEST_WORKERS`, execute a bounded search, and
record `GET /metrics` before and after it. Compare `jobsPerMinute`, queue wait
and execution p50/p95, and live queue state. Scale only
while those measurements justify it; more workers can eventually increase
database contention instead of improving throughput.

The runner now performs one discarded warm-up and calculates these three-run
medians and speedup directly, preventing a presenter from accidentally
cherry-picking the fastest single run.
