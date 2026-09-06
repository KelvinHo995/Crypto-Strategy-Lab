package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

func main() {
	candidateCount := flag.Int("candidates", 1000, "number of backtests in each run")
	candleCount := flag.Int("candles", 2000, "historical candles per backtest")
	workerList := flag.String("workers", "1,3", "comma-separated worker counts")
	repeatCount := flag.Int("repeat", 3, "measured repetitions per worker count")
	flag.Parse()
	if *candidateCount < 1 || *candleCount < experiment.MinCandlesForBacktest || *repeatCount < 1 {
		panic(fmt.Sprintf("candidates must be positive and candles must be at least %d", experiment.MinCandlesForBacktest))
	}

	workers, err := parseWorkers(*workerList)
	if err != nil {
		panic(err)
	}
	candles := fixtureCandles(*candleCount)
	fmt.Printf("candidates=%d candles_per_candidate=%d repeats=%d gomaxprocs=%d\n", *candidateCount, *candleCount, *repeatCount, runtime.GOMAXPROCS(0))
	var baseline time.Duration
	for _, workerCount := range workers {
		// Discard one warm-up so compilation, allocation and scheduler startup
		// noise are less likely to dominate the short controlled workload.
		measure(workerCount, *candidateCount, candles)
		durations := make([]time.Duration, 0, *repeatCount)
		for run := 1; run <= *repeatCount; run++ {
			duration := measure(workerCount, *candidateCount, candles)
			durations = append(durations, duration)
			fmt.Printf("workers=%d run=%d duration=%s throughput=%.2f jobs/s\n", workerCount, run, duration.Round(time.Millisecond), float64(*candidateCount)/duration.Seconds())
		}
		median := medianDuration(durations)
		if baseline == 0 {
			baseline = median
		}
		fmt.Printf("summary workers=%d median_duration=%s median_throughput=%.2f jobs/s speedup=%.2fx\n", workerCount, median.Round(time.Millisecond), float64(*candidateCount)/median.Seconds(), float64(baseline)/float64(median))
	}
}

func measure(workerCount, candidateCount int, candles []market.Candle) time.Duration {
	registry := strategy.NewRegistry()
	if err := registry.RegisterPlugin(strategy.NewMAPlugin(strategy.NewMAStrategy(5, 20))); err != nil {
		panic(err)
	}
	repo := newMemoryRepository()
	queue := experiment.NewInMemoryQueue(candidateCount)
	pool := experiment.NewWorkerPool(queue, registry, repo, workerCount)
	finished := make(chan string, candidateCount)
	pool.AddObserver(func(result experiment.Result) {
		if result.Status == "COMPLETED" || result.Status == "FAILED" {
			finished <- result.Status
		}
	})

	for i := 0; i < candidateCount; i++ {
		id := fmt.Sprintf("perf-%d-%d", workerCount, i)
		job := experiment.BacktestJob{
			ID: id, SearchID: "performance-proof", SearchTotal: candidateCount,
			Candidate: strategy.CandidateStrategy{ID: id, Policy: "majority", Instances: []strategy.StrategyInstance{{Type: "MA", Params: map[string]any{"maShortWindow": 5, "maLongWindow": 20}}}},
			Pair:      "BTCUSDT", Timeframe: "1m", Candles: candles, EnqueuedAt: time.Now().UnixMilli(),
			Config: experiment.Config{Pair: "BTCUSDT", StartingCapital: 1000, PositionSizePct: 1, AllowShort: true},
		}
		if err := queue.Enqueue(context.Background(), job); err != nil {
			panic(err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	started := time.Now()
	pool.Start(ctx)
	failed := 0
	for i := 0; i < candidateCount; i++ {
		if <-finished == "FAILED" {
			failed++
		}
	}
	duration := time.Since(started)
	cancel()
	pool.Wait()
	if failed > 0 {
		panic(fmt.Sprintf("performance run produced %d failed jobs", failed))
	}
	return duration
}

func medianDuration(values []time.Duration) time.Duration {
	ordered := append([]time.Duration(nil), values...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	mid := len(ordered) / 2
	if len(ordered)%2 == 1 {
		return ordered[mid]
	}
	return (ordered[mid-1] + ordered[mid]) / 2
}

func parseWorkers(raw string) ([]int, error) {
	parts := strings.Split(raw, ",")
	workers := make([]int, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || value < 1 || value > 64 {
			return nil, fmt.Errorf("invalid worker count %q; want 1-64", part)
		}
		workers = append(workers, value)
	}
	return workers, nil
}

func fixtureCandles(count int) []market.Candle {
	candles := make([]market.Candle, count)
	for i := range candles {
		closePrice := 100 + math.Sin(float64(i)/12)*8 + float64(i%37)/20
		candles[i] = market.Candle{Symbol: "BTCUSDT", Timeframe: "1m", OpenTime: int64(i) * 60_000, Open: closePrice - 0.2, High: closePrice + 0.5, Low: closePrice - 0.5, Close: closePrice, Volume: 10, IsClosed: true}
	}
	return candles
}

type memoryRepository struct {
	mu      sync.Mutex
	results map[string]experiment.Result
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{results: make(map[string]experiment.Result)}
}

func (r *memoryRepository) Save(_ context.Context, result experiment.Result) error {
	r.mu.Lock()
	r.results[result.ID] = result
	r.mu.Unlock()
	return nil
}

func (r *memoryRepository) Get(_ context.Context, id string) (experiment.Result, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result, ok := r.results[id]
	if !ok {
		return experiment.Result{}, experiment.ErrNotFound
	}
	return result, nil
}

func (r *memoryRepository) List(context.Context) ([]experiment.Result, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	results := make([]experiment.Result, 0, len(r.results))
	for _, result := range r.results {
		results = append(results, result)
	}
	return results, nil
}

func (r *memoryRepository) ListBySearch(ctx context.Context, searchID string) ([]experiment.Result, error) {
	results, _ := r.List(ctx)
	filtered := results[:0]
	for _, result := range results {
		if result.SearchID == searchID {
			filtered = append(filtered, result)
		}
	}
	return filtered, nil
}

func (*memoryRepository) SaveTrades(context.Context, string, []experiment.Trade) error { return nil }
func (*memoryRepository) ListTrades(context.Context, string) ([]experiment.Trade, error) {
	return []experiment.Trade{}, nil
}
