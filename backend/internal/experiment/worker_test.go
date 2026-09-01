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
	saves chan experiment.Result
}

type holdStrategy struct{}

func (holdStrategy) Name() string                            { return "Hold" }
func (holdStrategy) Analyze([]market.Candle) strategy.Signal { return strategy.Hold }

func newRecordingRepo() *recordingRepo {
	return &recordingRepo{saves: make(chan experiment.Result, 10)}
}

func (r *recordingRepo) Save(_ context.Context, res experiment.Result) error {
	r.saves <- res
	return nil
}
func (r *recordingRepo) Get(context.Context, string) (experiment.Result, error) {
	return experiment.Result{}, experiment.ErrNotFound
}
func (r *recordingRepo) List(context.Context) ([]experiment.Result, error) { return nil, nil }

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
			ID: "cand-1", Strategies: []string{"Evil"}, Policy: "majority",
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
				ID: "candidate-" + id, Strategies: []string{"Hold"}, Policy: "majority",
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
