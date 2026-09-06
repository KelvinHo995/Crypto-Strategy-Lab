package experiment

import (
	"context"
	"errors"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

var ErrQueueClosed = errors.New("experiment queue closed")

type BacktestJob struct {
	ID               string                     `json:"id"`
	SearchID         string                     `json:"searchId"`
	SearchTotal      int                        `json:"searchTotal"`
	Candidate        strategy.CandidateStrategy `json:"candidate"`
	Pair             string                     `json:"pair"`
	Timeframe        string                     `json:"timeframe"`
	From             int64                      `json:"from"`
	To               int64                      `json:"to"`
	Candles          []market.Candle            `json:"-"`
	Config           Config                     `json:"config"`
	DatasetPeriod    string                     `json:"datasetPeriod"`
	StrategyVersions map[string]string          `json:"strategyVersions"`
	EnqueuedAt       int64                      `json:"enqueuedAt"`
}

type Queue interface {
	Enqueue(context.Context, BacktestJob) error
	Dequeue(context.Context) (BacktestJob, error)
	Ack(context.Context, string) error
	Nack(context.Context, string, error) error
}

type QueueStats struct {
	Queued  int `json:"queued"`
	Running int `json:"running"`
	Failed  int `json:"failed"`
}

type QueueStatsProvider interface {
	Stats(context.Context) (QueueStats, error)
}

// PendingQueue persists the PENDING experiment row and queue message in one
// transaction, closing the crash window between the two writes.
type PendingQueue interface {
	EnqueuePending(context.Context, BacktestJob, Result) error
}

type LeaseQueue interface {
	Heartbeat(context.Context, string) error
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

func (q *InMemoryQueue) Ack(context.Context, string) error         { return nil }
func (q *InMemoryQueue) Nack(context.Context, string, error) error { return nil }

func (q *InMemoryQueue) Stats(context.Context) (QueueStats, error) {
	return QueueStats{Queued: len(q.jobs)}, nil
}

const (
	DefaultQueuePollInterval = time.Second
	DefaultQueueLease        = 15 * time.Minute
	DefaultQueueMaxAttempts  = 3
)
