package experiment

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

type WorkerPool struct {
	queue    Queue
	registry *strategy.Registry
	repo     Repository
	workers  int
	wg       sync.WaitGroup
	observer func(Result)
}

func (p *WorkerPool) SetObserver(observer func(Result)) { p.observer = observer }
func (p *WorkerPool) notify(r Result) {
	if p.observer != nil {
		p.observer(r)
	}
}

func NewWorkerPool(queue Queue, registry *strategy.Registry, repo Repository, workers int) *WorkerPool {
	if workers < 1 {
		workers = 1
	}
	return &WorkerPool{queue: queue, registry: registry, repo: repo, workers: workers}
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
	defer func() {
		if r := recover(); r != nil {
			log.Printf("experiment worker: panic on job %s: %v", job.ID, r)
			result := resultFromJob(job, "FAILED")
			if err := p.repo.Save(ctx, result); err != nil {
				log.Printf("experiment worker: save FAILED after panic for %s: %v", job.ID, err)
			}
			p.notify(result)
		}
	}()

	result := resultFromJob(job, "RUNNING")
	if err := p.repo.Save(ctx, result); err != nil {
		log.Printf("experiment worker: save RUNNING for %s: %v", job.ID, err)
		return
	}
	p.notify(result)
	combined, err := strategy.BuildFromCandidate(p.registry, job.Candidate)
	if err != nil {
		result.Status = "FAILED"
		if saveErr := p.repo.Save(ctx, result); saveErr != nil {
			log.Printf("experiment worker: save FAILED for %s: %v", job.ID, saveErr)
		}
		p.notify(result)
		return
	}
	trades := NewBacktester(job.Config).Run(combined, job.Candles)
	metrics := (Evaluator{StartingCapital: job.Config.StartingCapital}).Evaluate(trades)
	result.Return, result.MDD = metrics.Return, metrics.MDD
	result.TradeCount, result.WinRate = metrics.TradeCount, metrics.WinRate
	result.Wins, result.Losses = metrics.Wins, metrics.Losses
	result.TotalProfit, result.Status = metrics.TotalProfit, "COMPLETED"
	if err := p.repo.Save(ctx, result); err != nil {
		log.Printf("experiment worker: save COMPLETED for %s: %v", job.ID, err)
	}
	p.notify(result)
}

func resultFromJob(job BacktestJob, status string) Result {
	return Result{ID: job.ID, CandidateID: job.Candidate.ID,
		Strategies: append([]string(nil), job.Candidate.Strategies...),
		Params:     cloneMap(job.Candidate.Params), Policy: job.Candidate.Policy,
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

func DefaultStrategyVersions(names []string) map[string]string {
	versions := make(map[string]string, len(names))
	for _, name := range names {
		versions[name] = "v1"
	}
	return versions
}

func NewJobID(now time.Time) string { return fmt.Sprintf("exp-%d", now.UnixNano()) }
