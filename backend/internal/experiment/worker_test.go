package experiment_test

import (
	"context"
	"testing"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

type recordingRepo struct {
	saves      chan experiment.Result
	tradeSaves chan []experiment.Trade
}

type holdStrategy struct{}

func (holdStrategy) Name() string                            { return "Hold" }
func (holdStrategy) Analyze([]market.Candle) strategy.Signal { return strategy.Hold }

func newRecordingRepo() *recordingRepo {
	return &recordingRepo{saves: make(chan experiment.Result, 10), tradeSaves: make(chan []experiment.Trade, 10)}
}

func (r *recordingRepo) Save(_ context.Context, res experiment.Result) error {
	r.saves <- res
	return nil
}
func (r *recordingRepo) Get(context.Context, string) (experiment.Result, error) {
	return experiment.Result{}, experiment.ErrNotFound
}
func (r *recordingRepo) List(context.Context) ([]experiment.Result, error) { return nil, nil }
func (r *recordingRepo) ListBySearch(context.Context, string) ([]experiment.Result, error) {
	return nil, nil
}
func (r *recordingRepo) SaveTrades(_ context.Context, _ string, trades []experiment.Trade) error {
	r.tradeSaves <- trades
	return nil
}
func (r *recordingRepo) ListTrades(context.Context, string) ([]experiment.Trade, error) {
	return nil, nil
}

// A malformed/buggy strategy factory (e.g. a bad third-party plugin) panics
// instead of returning an error. The worker pool must survive this — see
// the recover() in WorkerPool.run.
func TestWorkerPoolRun_RecoversPanicAndMarksFailed(t *testing.T) {
	registry := strategy.NewRegistry()
	registry.RegisterFactory("Evil", func(map[string]any) strategy.Strategy {
		panic("boom")
	})

	repo := newRecordingRepo()
	queue := experiment.NewInMemoryQueue(1)
	pool := experiment.NewWorkerPool(queue, registry, repo, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	job := experiment.BacktestJob{
		ID: "job-panic",
		Candidate: strategy.CandidateStrategy{
			ID: "cand-1", Instances: []strategy.StrategyInstance{{Type: "Evil"}}, Policy: "majority",
		},
		Candles: []market.Candle{{Symbol: "BTCUSDT", OpenTime: 1, Open: 100, High: 101, Low: 99, Close: 100}},
		Config: experiment.Config{
			Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
			StopLossPct: 0.02, TakeProfitPct: 0.04, FeePct: 0.001, SlippageBps: 5, Window: 1,
		},
		EnqueuedAt: time.Now().UnixMilli(),
	}
	if err := queue.Enqueue(ctx, job); err != nil {
		t.Fatal(err)
	}

	var last experiment.Result
	for i := 0; i < 2; i++ {
		select {
		case last = <-repo.saves:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for worker to save a result")
		}
	}

	if last.Status != "FAILED" {
		t.Fatalf("Status = %q, want FAILED", last.Status)
	}
	if last.ID != job.ID {
		t.Fatalf("ID = %q, want %q", last.ID, job.ID)
	}
}

func TestWorkerPoolRun_ObserverPanicDoesNotStopWorker(t *testing.T) {
	registry := strategy.NewRegistry()
	registry.RegisterFactory("Hold", func(map[string]any) strategy.Strategy { return holdStrategy{} })
	repo := newRecordingRepo()
	queue := experiment.NewInMemoryQueue(2)
	pool := experiment.NewWorkerPool(queue, registry, repo, 1)
	pool.SetObserver(func(experiment.Result) { panic("observer boom") })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	for _, id := range []string{"job-observer-1", "job-observer-2"} {
		job := experiment.BacktestJob{
			ID: id,
			Candidate: strategy.CandidateStrategy{
				ID: "candidate-" + id, Instances: []strategy.StrategyInstance{{Type: "Hold"}}, Policy: "majority",
			},
			Candles: []market.Candle{{Symbol: "BTCUSDT", OpenTime: 1, Open: 100, High: 101, Low: 99, Close: 100}},
			Config: experiment.Config{
				Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
				StopLossPct: 0.02, TakeProfitPct: 0.04, FeePct: 0.001, SlippageBps: 5, Window: 1,
			},
			EnqueuedAt: time.Now().UnixMilli(),
		}
		if err := queue.Enqueue(ctx, job); err != nil {
			t.Fatal(err)
		}
	}

	completed := map[string]bool{}
	deadline := time.After(3 * time.Second)
	for len(completed) < 2 {
		select {
		case result := <-repo.saves:
			if result.Status == "COMPLETED" {
				completed[result.ID] = true
			}
		case <-deadline:
			t.Fatalf("worker stopped after observer panic; completed jobs: %v", completed)
		}
	}
}

func TestWorkerPoolAddObserver_CoexistsAndUnsubscribes(t *testing.T) {
	registry := strategy.NewRegistry()
	registry.RegisterFactory("Hold", func(map[string]any) strategy.Strategy { return holdStrategy{} })
	repo := newRecordingRepo()
	queue := experiment.NewInMemoryQueue(2)
	pool := experiment.NewWorkerPool(queue, registry, repo, 1)

	setCalls := make(chan string, 10)
	pool.SetObserver(func(r experiment.Result) { setCalls <- r.ID })

	addCalls := make(chan string, 10)
	unsubscribe := pool.AddObserver(func(r experiment.Result) { addCalls <- r.ID })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	enqueue := func(id string) {
		job := experiment.BacktestJob{
			ID: id,
			Candidate: strategy.CandidateStrategy{
				ID: "candidate-" + id, Instances: []strategy.StrategyInstance{{Type: "Hold"}}, Policy: "majority",
			},
			Candles: []market.Candle{{Symbol: "BTCUSDT", OpenTime: 1, Open: 100, High: 101, Low: 99, Close: 100}},
			Config: experiment.Config{
				Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
				StopLossPct: 0.02, TakeProfitPct: 0.04, FeePct: 0.001, SlippageBps: 5, Window: 1,
			},
			EnqueuedAt: time.Now().UnixMilli(),
		}
		if err := queue.Enqueue(ctx, job); err != nil {
			t.Fatal(err)
		}
	}

	enqueue("job-both-1")
	// notify() fires once for RUNNING and once for COMPLETED — drain both
	// from each observer before unsubscribing, or a buffered COMPLETED
	// message left over from before unsubscribe() would look like a leak.
	waitForID(t, setCalls, "job-both-1")
	waitForID(t, setCalls, "job-both-1")
	waitForID(t, addCalls, "job-both-1")
	waitForID(t, addCalls, "job-both-1")

	unsubscribe()
	enqueue("job-set-only")
	waitForID(t, setCalls, "job-set-only")

	select {
	case id := <-addCalls:
		t.Fatalf("unsubscribed observer still received %q", id)
	case <-time.After(200 * time.Millisecond):
	}
}

// nackTrackingQueue wraps InMemoryQueue to observe Nack calls — the plain
// InMemoryQueue's Nack is a no-op, so it can't distinguish "retried" from
// "silently dropped" the way PostgresQueue's real retry semantics would.
type nackTrackingQueue struct {
	*experiment.InMemoryQueue
	nacked chan string
}

func (q *nackTrackingQueue) Nack(ctx context.Context, id string, cause error) error {
	q.nacked <- id
	return q.InMemoryQueue.Nack(ctx, id, cause)
}

type emptyCandleRepo struct{}

func (emptyCandleRepo) Upsert(context.Context, []market.Candle) error { return nil }
func (emptyCandleRepo) Range(context.Context, string, string, int64, int64) ([]market.Candle, error) {
	return nil, nil
}

// A job with no embedded candles falls back to fetching them on demand. Too
// few candles to backtest against should be retried, not failed outright —
// the shortfall might just mean backfill for this range hasn't landed yet,
// same reasoning as every other failure path in run().
func TestWorkerPoolRun_RetriesInsufficientCandlesInsteadOfFailingOutright(t *testing.T) {
	registry := strategy.NewRegistry()
	registry.RegisterFactory("Hold", func(map[string]any) strategy.Strategy { return holdStrategy{} })
	repo := newRecordingRepo()
	queue := &nackTrackingQueue{InMemoryQueue: experiment.NewInMemoryQueue(1), nacked: make(chan string, 10)}
	pool := experiment.NewWorkerPool(queue, registry, repo, 1)
	pool.SetCandleRepository(emptyCandleRepo{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	job := experiment.BacktestJob{
		ID:        "job-thin-data",
		Pair:      "BTCUSDT",
		Timeframe: "1h",
		From:      1,
		To:        2,
		Candidate: strategy.CandidateStrategy{ID: "candidate-thin", Instances: []strategy.StrategyInstance{{Type: "Hold"}}, Policy: "majority"},
		Config: experiment.Config{
			Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
			StopLossPct: 0.02, TakeProfitPct: 0.04, FeePct: 0.001, SlippageBps: 5, Window: 1,
		},
		EnqueuedAt: time.Now().UnixMilli(),
	}
	if err := queue.Enqueue(ctx, job); err != nil {
		t.Fatal(err)
	}

	select {
	case id := <-queue.nacked:
		if id != job.ID {
			t.Fatalf("nacked job = %q, want %q", id, job.ID)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for insufficient-candle job to be retried")
	}

	select {
	case result := <-repo.saves:
		t.Fatalf("insufficient-candle job should retry without saving a result first, got %+v", result)
	case <-time.After(200 * time.Millisecond):
	}
}

func waitForID(t *testing.T, ch chan string, want string) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case id := <-ch:
			if id == want {
				return
			}
		case <-deadline:
			t.Fatalf("timed out waiting for %q", want)
		}
	}
}

// The bug this reproduces: job.Config.Window was a fixed guess set at job
// construction time, unrelated to what the actual resolved strategy needs.
// MA with default factory params (20/50) needs 51 candles per call but was
// only ever handed 20 (or, here, a deliberately worse 5) — so it silently
// returned Hold forever and every backtest using default MA completed with
// zero trades. The worker must now size the window from the real resolved
// strategy's MinLookback(), ignoring whatever job.Config.Window says.
func TestWorkerPoolRun_SizesWindowFromResolvedStrategyNotJobConfig(t *testing.T) {
	registry := strategy.NewRegistry()
	registry.RegisterFactory("MA", strategy.MAFactory) // default 20/50, no custom params supplied
	repo := newRecordingRepo()
	queue := experiment.NewInMemoryQueue(1)
	pool := experiment.NewWorkerPool(queue, registry, repo, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	// Flat, then a sharp jump — a straight ramp never actually "crosses"
	// (short MA sits above long MA from the first computable point), so this
	// mirrors the same flat-then-jump pattern TestMAStrategy_Analyze uses to
	// force a real, discrete crossover event.
	candles := make([]market.Candle, 80)
	for i := range candles {
		price := 100.0
		if i >= 60 {
			price = 100.0 + float64(i-59)*5
		}
		candles[i] = market.Candle{Symbol: "BTCUSDT", OpenTime: int64(i), Open: price, High: price + 1, Low: price - 1, Close: price, Volume: 1}
	}

	job := experiment.BacktestJob{
		ID:        "job-window-sizing",
		Candidate: strategy.CandidateStrategy{ID: "cand-ma-default", Instances: []strategy.StrategyInstance{{Type: "MA"}}, Policy: "majority"},
		Candles:   candles,
		Config: experiment.Config{
			Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
			StopLossPct: 0.02, TakeProfitPct: 0.04, FeePct: 0.001, SlippageBps: 5,
			Window: 5, // deliberately far too small for MA's real 51-candle need
		},
		EnqueuedAt: time.Now().UnixMilli(),
	}
	if err := queue.Enqueue(ctx, job); err != nil {
		t.Fatal(err)
	}

	var last experiment.Result
	for i := 0; i < 2; i++ {
		select {
		case last = <-repo.saves:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for worker to save a result")
		}
	}

	if last.Status != "COMPLETED" {
		t.Fatalf("Status = %q, want COMPLETED", last.Status)
	}
	if last.TradeCount == 0 {
		t.Fatal("MA produced zero trades — worker trusted job.Config.Window instead of sizing it from the resolved strategy's MinLookback()")
	}
}

func TestWorkerPoolRun_PersistsTradesForCompletedBacktest(t *testing.T) {
	registry := strategy.NewRegistry()
	registry.RegisterFactory("MA", strategy.MAFactory)
	repo := newRecordingRepo()
	queue := experiment.NewInMemoryQueue(1)
	pool := experiment.NewWorkerPool(queue, registry, repo, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	candles := make([]market.Candle, 80)
	for i := range candles {
		price := 100.0
		if i >= 60 {
			price = 100.0 + float64(i-59)*5
		}
		candles[i] = market.Candle{Symbol: "BTCUSDT", OpenTime: int64(i), Open: price, High: price + 1, Low: price - 1, Close: price, Volume: 1}
	}

	job := experiment.BacktestJob{
		ID:        "job-trade-persistence",
		Candidate: strategy.CandidateStrategy{ID: "cand-ma-default", Instances: []strategy.StrategyInstance{{Type: "MA"}}, Policy: "majority"},
		Candles:   candles,
		Config: experiment.Config{
			Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1,
			StopLossPct: 0.02, TakeProfitPct: 0.04, FeePct: 0.001, SlippageBps: 5,
		},
		EnqueuedAt: time.Now().UnixMilli(),
	}
	if err := queue.Enqueue(ctx, job); err != nil {
		t.Fatal(err)
	}

	var last experiment.Result
	for i := 0; i < 2; i++ {
		select {
		case last = <-repo.saves:
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for worker to save a result")
		}
	}
	if last.Status != "COMPLETED" || last.TradeCount == 0 {
		t.Fatalf("result = %+v, want a COMPLETED result with trades", last)
	}

	select {
	case trades := <-repo.tradeSaves:
		if len(trades) != last.TradeCount {
			t.Fatalf("SaveTrades got %d trades, want %d (matching Result.TradeCount)", len(trades), last.TradeCount)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the worker to call SaveTrades")
	}
}
