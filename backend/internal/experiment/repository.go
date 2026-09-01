package experiment

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("experiment not found")

type Repository interface {
	Save(ctx context.Context, r Result) error
	Get(ctx context.Context, id string) (Result, error)
	List(ctx context.Context) ([]Result, error)
	// ListBySearch returns every Result sharing searchID, for progress
	// tracking and per-search no-improvement checks.
	ListBySearch(ctx context.Context, searchID string) ([]Result, error)
	// MarkStaleRunningFailed atomically fails RUNNING rows untouched for
	// olderThan and returns the ones it changed. Implementations must be
	// safe for multiple instances to call concurrently without duplicating
	// work — see PostgresRepository's advisory-lock guard.
	MarkStaleRunningFailed(ctx context.Context, olderThan time.Duration) ([]Result, error)
}
