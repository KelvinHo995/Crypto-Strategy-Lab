package experiment

import (
	"context"
	"log"
	"time"
)

const (
	DefaultStaleThreshold = 15 * time.Minute
	DefaultSweepInterval  = 1 * time.Minute
)

type Sweeper struct{}

func NewSweeper() *Sweeper { return &Sweeper{} }

func (s *Sweeper) Sweep(ctx context.Context, repo Repository, olderThan time.Duration, notify func(Result)) {
	results, err := repo.MarkStaleRunningFailed(ctx, olderThan)
	if err != nil {
		log.Printf("stale sweep: %v", err)
		return
	}
	for _, r := range results {
		log.Printf("stale sweep: %s RUNNING for >= %s, marked FAILED", r.ID, olderThan)
		if notify != nil {
			notify(r)
		}
	}
}

// Run calls Sweep on a fixed interval until ctx is done.
func (s *Sweeper) Run(ctx context.Context, repo Repository, olderThan, interval time.Duration, notify func(Result)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.Sweep(ctx, repo, olderThan, notify)
		}
	}
}
