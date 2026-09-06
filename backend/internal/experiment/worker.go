package experiment

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
	"github.com/google/uuid"
)

type WorkerPool struct {
	queue     Queue
	registry  *strategy.Registry
	repo      Repository
	candles   market.CandleRepository
	workers   int
	wg        sync.WaitGroup
	obsMu     sync.Mutex
	observers map[int]func(Result)
	nextObsID int
}

// SetObserver replaces every existing observer with just this one — kept
// for the router's single always-on WebSocket-hub wiring. Use AddObserver
// for anything that needs to coexist with it (e.g. a search loop tracking
// its own candidates' completions).
func (p *WorkerPool) SetObserver(observer func(Result)) {
	p.obsMu.Lock()
	defer p.obsMu.Unlock()
	p.observers = map[int]func(Result){0: observer}
	p.nextObsID = 1
}

// AddObserver registers an additional observer without disturbing any
// already registered. The returned func removes it; callers that only run
// for a while (like a search loop) must call it when done, or the closure
// leaks for the life of the process.
func (p *WorkerPool) AddObserver(observer func(Result)) (unsubscribe func()) {
	p.obsMu.Lock()
	defer p.obsMu.Unlock()
	if p.observers == nil {
		p.observers = map[int]func(Result){}
	}
	id := p.nextObsID
	p.nextObsID++
	p.observers[id] = observer
	return func() {
		p.obsMu.Lock()
		defer p.obsMu.Unlock()
		delete(p.observers, id)
	}
}

func (p *WorkerPool) notify(r Result) {
	p.obsMu.Lock()
	observers := make([]func(Result), 0, len(p.observers))
	for _, o := range p.observers {
		observers = append(observers, o)
	}
	p.obsMu.Unlock()
	for _, observer := range observers {
		p.notifyOne(observer, r)
	}
}

func (p *WorkerPool) notifyOne(observer func(Result), r Result) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("experiment worker: observer panic for %s: %v", r.ID, recovered)
		}
	}()
	observer(r)
}

func NewWorkerPool(queue Queue, registry *strategy.Registry, repo Repository, workers int) *WorkerPool {
	if workers < 1 {
		workers = 1
	}
	return &WorkerPool{queue: queue, registry: registry, repo: repo, workers: workers}
}

func (p *WorkerPool) SetCandleRepository(repository market.CandleRepository) {
	p.candles = repository
}

func (p *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				job, err := p.queue.Dequeue(ctx)
				if err != nil {
					return
				}
				p.run(ctx, job)
			}
		}()
	}
}

func (p *WorkerPool) Wait() { p.wg.Wait() }

// run must not let a panic escape: it runs on a long-lived pool goroutine
// with no recover() above it, so an unrecovered panic here would crash the
// whole server process, not just fail this one job.
func (p *WorkerPool) run(ctx context.Context, job BacktestJob) {
	var resolvedStrategy strategy.Strategy
	defer func() {
		if r := recover(); r != nil {
			log.Printf("experiment worker: panic on job %s: %v", job.ID, r)
			result := resultFromJob(job, "FAILED")
			result.SentimentModels = sentimentModelsFrom(resolvedStrategy)
			if err := p.repo.Save(ctx, result); err != nil {
				log.Printf("experiment worker: save FAILED after panic for %s: %v", job.ID, err)
				p.retry(ctx, job.ID, err)
				return
			}
			p.notify(result)
			p.ack(ctx, job.ID)
		}
	}()

	candles := job.Candles
	if len(candles) == 0 && p.candles != nil {
		var err error
		candles, err = p.candles.Range(ctx, job.Pair, job.Timeframe, job.From, job.To)
		if err != nil {
			p.retry(ctx, job.ID, fmt.Errorf("load historical candles: %w", err))
			return
		}
		if len(candles) < MinCandlesForBacktest {
			// Retried like every other failure path here: a short count can
			// mean the range genuinely has no data (permanent — exhausts
			// retries and the queue marks it FAILED on its own) or that
			// backfill for this range just hasn't landed yet (transient —
			// a retry a few seconds later succeeds).
			p.retry(ctx, job.ID, fmt.Errorf("only %d candles available, want at least %d", len(candles), MinCandlesForBacktest))
			return
		}
	}

	result := resultFromJob(job, "RUNNING")
	if err := p.repo.Save(ctx, result); err != nil {
		log.Printf("experiment worker: save RUNNING for %s: %v", job.ID, err)
		p.retry(ctx, job.ID, err)
		return
	}
	stopHeartbeat := p.startHeartbeat(ctx, job.ID)
	defer stopHeartbeat()
	p.notify(result)
	combined, err := strategy.BuildFromCandidate(p.registry, job.Candidate)
	if err != nil {
		result.Status = "FAILED"
		if saveErr := p.repo.Save(ctx, result); saveErr != nil {
			log.Printf("experiment worker: save FAILED for %s: %v", job.ID, saveErr)
			p.retry(ctx, job.ID, saveErr)
			return
		}
		p.notify(result)
		p.ack(ctx, job.ID)
		return
	}
	resolvedStrategy = combined
	config := job.Config
	if la, ok := combined.(strategy.LookbackAware); ok {
		config.Window = la.MinLookback()
	} else if config.Window <= 0 {
		config.Window = 1
	}
	trades := NewBacktester(config).Run(combined, candles)
	result.SentimentModels = sentimentModelsFrom(combined)
	metrics := (Evaluator{StartingCapital: job.Config.StartingCapital}).Evaluate(trades)
	result.Return, result.MDD = metrics.Return, metrics.MDD
	result.TradeCount, result.WinRate = metrics.TradeCount, metrics.WinRate
	result.Wins, result.Losses = metrics.Wins, metrics.Losses
	result.TotalProfit, result.Status = metrics.TotalProfit, "COMPLETED"
	// Trades are saved before the result flips to COMPLETED. A run is not
	// reproducible if its advertised trade detail could not be persisted, so
	// fail the result explicitly instead of publishing a misleading success.
	if err := p.repo.SaveTrades(ctx, job.ID, trades); err != nil {
		log.Printf("experiment worker: save trades for %s: %v", job.ID, err)
		result.Status = "FAILED"
		if saveErr := p.repo.Save(ctx, result); saveErr != nil {
			log.Printf("experiment worker: save FAILED after trade persistence error for %s: %v", job.ID, saveErr)
			p.retry(ctx, job.ID, saveErr)
			return
		}
		p.notify(result)
		p.ack(ctx, job.ID)
		return
	}
	if err := p.repo.Save(ctx, result); err != nil {
		log.Printf("experiment worker: save COMPLETED for %s: %v", job.ID, err)
		p.retry(ctx, job.ID, err)
		return
	}
	p.notify(result)
	p.ack(ctx, job.ID)
}

func sentimentModelsFrom(resolved strategy.Strategy) []strategy.SentimentModelIdentity {
	if provider, ok := resolved.(strategy.SentimentModelProvider); ok {
		return provider.SentimentModels()
	}
	return nil
}

func (p *WorkerPool) ack(ctx context.Context, id string) {
	if err := p.queue.Ack(ctx, id); err != nil {
		log.Printf("experiment worker: ack %s: %v", id, err)
	}
}

func (p *WorkerPool) retry(ctx context.Context, id string, cause error) {
	if err := p.queue.Nack(ctx, id, cause); err != nil {
		log.Printf("experiment worker: nack %s: %v", id, err)
	}
}

func (p *WorkerPool) startHeartbeat(ctx context.Context, id string) func() {
	queue, ok := p.queue.(LeaseQueue)
	if !ok {
		return func() {}
	}
	heartbeatCtx, cancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-heartbeatCtx.Done():
				return
			case <-ticker.C:
				if err := queue.Heartbeat(heartbeatCtx, id); err != nil {
					log.Printf("experiment worker: heartbeat %s: %v", id, err)
				}
			}
		}
	}()
	return cancel
}

func resultFromJob(job BacktestJob, status string) Result {
	searchID := job.SearchID
	if searchID == "" {
		searchID = job.ID
	}
	searchTotal := job.SearchTotal
	if searchTotal < 1 {
		searchTotal = 1
	}
	return Result{ID: job.ID, SearchID: searchID, SearchTotal: searchTotal, CandidateID: job.Candidate.ID,
		Pair: job.Pair, Timeframe: job.Timeframe,
		Instances: cloneInstances(job.Candidate.Instances), Policy: job.Candidate.Policy,
		StrategyVersions: cloneStringMap(job.StrategyVersions), DatasetPeriod: job.DatasetPeriod,
		Status: status, CreatedAt: job.EnqueuedAt}
}

func cloneMap(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
func cloneStringMap(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneInstances(src []strategy.StrategyInstance) []strategy.StrategyInstance {
	dst := make([]strategy.StrategyInstance, len(src))
	for i, inst := range src {
		dst[i] = strategy.StrategyInstance{Type: inst.Type, Params: cloneMap(inst.Params), Weight: inst.Weight}
	}
	return dst
}

// UUID suffix: timestamp alone can collide across a tight generation loop or across instances.
func NewJobID(now time.Time) string {
	return fmt.Sprintf("exp-%d-%s", now.UnixNano(), uuid.NewString())
}
