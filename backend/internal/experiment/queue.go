package experiment

import (
	"context"
	"errors"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

var ErrQueueClosed = errors.New("experiment queue closed")

type BacktestJob struct {
	ID               string
	Candidate        strategy.CandidateStrategy
	Candles          []market.Candle
	Config           Config
	DatasetPeriod    string
	StrategyVersions map[string]string
	EnqueuedAt       int64
}

type Queue interface {
	Enqueue(context.Context, BacktestJob) error
	Dequeue(context.Context) (BacktestJob, error)
}

type InMemoryQueue struct{ jobs chan BacktestJob }

func NewInMemoryQueue(capacity int) *InMemoryQueue {
	if capacity < 1 {
		capacity = 1
	}
	return &InMemoryQueue{jobs: make(chan BacktestJob, capacity)}
}

func (q *InMemoryQueue) Enqueue(ctx context.Context, job BacktestJob) error {
	select {
	case q.jobs <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *InMemoryQueue) Dequeue(ctx context.Context) (BacktestJob, error) {
	select {
	case job, ok := <-q.jobs:
		if !ok {
			return BacktestJob{}, ErrQueueClosed
		}
		return job, nil
	case <-ctx.Done():
		return BacktestJob{}, ctx.Err()
	}
}
