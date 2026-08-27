package experiment

import (
	"context"
	"fmt"
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

func (p *WorkerPool) run(ctx context.Context, job BacktestJob) {
	result := resultFromJob(job, "RUNNING")
	if err := p.repo.Save(ctx, result); err != nil {
		return
	}
	combined, err := strategy.BuildFromCandidate(p.registry, job.Candidate)
	if err != nil {
		result.Status = "FAILED"
		_ = p.repo.Save(ctx, result)
		return
	}
	trades := NewBacktester(job.Config).Run(combined, job.Candles)
	metrics := (Evaluator{StartingCapital: job.Config.StartingCapital}).Evaluate(trades)
	result.Return, result.MDD = metrics.Return, metrics.MDD
	result.TradeCount, result.WinRate = metrics.TradeCount, metrics.WinRate
	result.Wins, result.Losses = metrics.Wins, metrics.Losses
	result.TotalProfit, result.Status = metrics.TotalProfit, "COMPLETED"
	_ = p.repo.Save(ctx, result)
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
