package experiment

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

// observerRegistry is the slice of WorkerPool RunSearchLoop actually needs —
// narrowed so tests can simulate job completions without a real worker pool
// and backtest pipeline behind it.
type observerRegistry interface {
	AddObserver(func(Result)) (unsubscribe func())
}

type LoopParams struct {
	SearchID           string
	Pair               string
	TimeFrame          string
	From, To           int64
	StartingCapital    float64
	DatasetPeriod      string
	MaxCandidates      int
	MaxDuration        time.Duration
	NoImprovementLimit int
}

// RunSearchLoop generates candidates one at a time via gen, enqueueing each
// as a job, until MaxCandidates is reached or MaxDuration/NoImprovementLimit
// (whichever configured) trips first — OR semantics per ADR-0011. Already
// enqueued jobs keep running on the worker pool untouched; stopping only
// means no further candidates get generated.
func RunSearchLoop(ctx context.Context, params LoopParams, gen strategy.StrategyGenerator, pool observerRegistry, repo Repository, queue Queue) {
	seen := map[string]bool{}
	started := time.Now()

	var mu sync.Mutex
	bestScore := math.Inf(-1)
	stepsSinceImprovement := 0

	// The generator can race ahead of completions by up to the queue's
	// buffer size (workers take real time; enqueueing doesn't) — so this
	// stop can overshoot NoImprovementLimit by that much before it lands,
	// not stop at exactly N. Accepted: tightening it would mean waiting for
	// each candidate to finish before generating the next, serializing
	// generation to completion pace and starving the worker pool's
	// parallelism, which defeats the point of having 3 workers.
	if params.NoImprovementLimit > 0 {
		unsubscribe := pool.AddObserver(func(r Result) {
			if r.SearchID != params.SearchID || (r.Status != "COMPLETED" && r.Status != "FAILED") {
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if r.Status == "COMPLETED" {
				if s := Score(r); s > bestScore {
					bestScore, stepsSinceImprovement = s, 0
					return
				}
			}
			stepsSinceImprovement++
		})
		defer unsubscribe()
	}

	for i := 0; i < params.MaxCandidates; i++ {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if params.MaxDuration > 0 && time.Since(started) >= params.MaxDuration {
			return
		}
		if params.NoImprovementLimit > 0 {
			mu.Lock()
			stop := stepsSinceImprovement >= params.NoImprovementLimit
			mu.Unlock()
			if stop {
				return
			}
		}

		var candidate strategy.CandidateStrategy
		for attempt := 0; attempt < 10; attempt++ {
			candidate = gen.Generate()
			key := candidateKey(candidate)
			if !seen[key] {
				seen[key] = true
				break
			}
		}

		id := NewJobID(time.Now())
		versions := DefaultStrategyVersions(candidate.Strategies)
		now := time.Now()
		job := BacktestJob{
			ID: id, SearchID: params.SearchID, SearchTotal: params.MaxCandidates,
			Candidate: candidate,
			Pair:      params.Pair, Timeframe: params.TimeFrame, From: params.From, To: params.To,
			Config: Config{
				Pair: params.Pair, StartingCapital: params.StartingCapital,
				PositionSizePct: 1, StopLossPct: 0.02, TakeProfitPct: 0.04,
				FeePct: 0.001, SlippageBps: 5, Window: 20,
			},
			DatasetPeriod: params.DatasetPeriod, StrategyVersions: versions,
			EnqueuedAt: now.UnixMilli(),
		}
		pending := Result{
			ID: id, SearchID: params.SearchID, SearchTotal: params.MaxCandidates,
			CandidateID: candidate.ID, Strategies: candidate.Strategies,
			Params: candidate.Params, Policy: candidate.Policy,
			StrategyVersions: versions, DatasetPeriod: params.DatasetPeriod,
			Status: "PENDING", CreatedAt: now.UnixMilli(),
		}
		if err := repo.Save(ctx, pending); err != nil {
			log.Printf("search loop %s: save pending: %v", params.SearchID, err)
			continue
		}
		if err := queue.Enqueue(ctx, job); err != nil {
			log.Printf("search loop %s: enqueue: %v", params.SearchID, err)
			return
		}
	}
}

func candidateKey(c strategy.CandidateStrategy) string {
	strategies := append([]string(nil), c.Strategies...)
	sort.Strings(strategies)
	params, _ := json.Marshal(c.Params)
	return strings.Join(strategies, ",") + "|" + c.Policy + "|" + string(params)
}
