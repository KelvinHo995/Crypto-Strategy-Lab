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

// How often the generator re-checks whether an in-flight candidate has
// resolved while paced (see NoImprovementLimit handling below).
const searchLoopPaceInterval = 50 * time.Millisecond

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
	FeePct             float64 // e.g. 0.001 = 10bps; defaulted by the caller if unset
	SlippageBps        float64 // e.g. 5 = 5bps; defaulted by the caller if unset
	VersionResolver    strategy.PluginVersionResolver
}

// RunSearchLoop generates candidates one at a time via gen, enqueueing each
// as a job, until MaxCandidates is reached or MaxDuration/NoImprovementLimit
// (whichever configured) trips first — OR semantics per ADR-0011. Already
// enqueued jobs keep running on the worker pool untouched; stopping only
// means no further candidates get generated.
//
// onStopped, if non-nil, fires exactly once when generation ends early
// (MaxDuration/NoImprovementLimit/cancellation/enqueue failure) — not on the
// normal case of generating all MaxCandidates, which the existing
// tested==total progress mechanism already detects on its own once every
// candidate finishes. status is "FAILED" if nothing was ever enqueued,
// "STOPPED" if some candidates are still running.
func RunSearchLoop(ctx context.Context, params LoopParams, gen strategy.StrategyGenerator, pool observerRegistry, repo Repository, queue Queue, onStopped func(status, reason string)) {
	seen := map[string]bool{}
	started := time.Now()
	enqueued := 0
	reason := "max-candidates"

	var mu sync.Mutex
	bestScore := math.Inf(-1)
	stepsSinceImprovement := 0
	completed := 0

	// Without a cap, a fast generator against a queue with no backpressure
	// of its own (the production Postgres queue has none — see ADR-0011)
	// can enqueue every candidate before enough of them have actually
	// completed for stepsSinceImprovement to ever reach the limit, so
	// no-improvement would never fire early at all. Capping in-flight
	// candidates (enqueued minus completed) at NoImprovementLimit bounds
	// the overshoot to exactly that limit: you can't know whether the
	// last N had no improvement while more than N are still unresolved.
	// This does not serialize generation to one-at-a-time — up to
	// NoImprovementLimit candidates can still be running in parallel
	// across the worker pool at once.
	if params.NoImprovementLimit > 0 {
		unsubscribe := pool.AddObserver(func(r Result) {
			if r.SearchID != params.SearchID || (r.Status != "COMPLETED" && r.Status != "FAILED") {
				return
			}
			mu.Lock()
			defer mu.Unlock()
			completed++
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

generate:
	for i := 0; i < params.MaxCandidates; i++ {
		select {
		case <-ctx.Done():
			reason = "cancelled"
			break generate
		default:
		}
		if params.MaxDuration > 0 && time.Since(started) >= params.MaxDuration {
			reason = "max-duration"
			break generate
		}
		if params.NoImprovementLimit > 0 {
			mu.Lock()
			stop := stepsSinceImprovement >= params.NoImprovementLimit
			mu.Unlock()
			if stop {
				reason = "no-improvement"
				break generate
			}
			// Pace generation to the cap described above: don't generate
			// another candidate while NoImprovementLimit are already
			// in flight, wait for at least one to resolve first.
			for {
				mu.Lock()
				inFlight := enqueued - completed
				mu.Unlock()
				if inFlight < params.NoImprovementLimit {
					break
				}
				select {
				case <-ctx.Done():
					reason = "cancelled"
					break generate
				case <-time.After(searchLoopPaceInterval):
				}
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
		versions := map[string]string{}
		if params.VersionResolver != nil {
			var err error
			versions, err = params.VersionResolver.VersionsFor(candidate.Instances)
			if err != nil {
				log.Printf("search loop %s: resolve strategy versions: %v", params.SearchID, err)
				reason = "invalid-plugin"
				break generate
			}
		}
		now := time.Now()
		job := BacktestJob{
			ID: id, SearchID: params.SearchID, SearchTotal: params.MaxCandidates,
			Candidate: candidate,
			Pair:      params.Pair, Timeframe: params.TimeFrame, From: params.From, To: params.To,
			Config: Config{
				Pair: params.Pair, StartingCapital: params.StartingCapital,
				PositionSizePct: 1, StopLossPct: 0.02, TakeProfitPct: 0.04,
				FeePct: params.FeePct, SlippageBps: params.SlippageBps, AllowShort: true, // Window is sized per-candidate by the worker (see worker.go)
			},
			DatasetPeriod: params.DatasetPeriod, StrategyVersions: versions,
			EnqueuedAt: now.UnixMilli(),
		}
		pending := Result{
			ID: id, SearchID: params.SearchID, SearchTotal: params.MaxCandidates,
			CandidateID: candidate.ID, Instances: candidate.Instances, Policy: candidate.Policy,
			Pair: params.Pair, Timeframe: params.TimeFrame,
			StrategyVersions: versions, DatasetPeriod: params.DatasetPeriod,
			Status: "PENDING", CreatedAt: now.UnixMilli(),
		}

		if durable, ok := queue.(PendingQueue); ok {
			if err := durable.EnqueuePending(ctx, job, pending); err != nil {
				log.Printf("search loop %s: enqueue pending: %v", params.SearchID, err)
				reason = "enqueue-failed"
				break generate
			}
		} else {
			if err := repo.Save(ctx, pending); err != nil {
				log.Printf("search loop %s: save pending: %v", params.SearchID, err)
				continue
			}
			if err := queue.Enqueue(ctx, job); err != nil {
				log.Printf("search loop %s: enqueue: %v", params.SearchID, err)
				reason = "enqueue-failed"
				break generate
			}
		}
		enqueued++
	}

	if reason != "max-candidates" && onStopped != nil {
		status := "STOPPED"
		if enqueued == 0 {
			status = "FAILED"
		}
		onStopped(status, reason)
	}
}

func candidateKey(c strategy.CandidateStrategy) string {
	keys := make([]string, len(c.Instances))
	for i, inst := range c.Instances {
		data, _ := json.Marshal(inst)
		keys[i] = string(data)
	}
	sort.Strings(keys)
	return c.Policy + "|" + strings.Join(keys, ",")
}
