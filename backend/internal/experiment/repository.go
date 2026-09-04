package experiment

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("experiment not found")

type Repository interface {
	Save(ctx context.Context, r Result) error
	Get(ctx context.Context, id string) (Result, error)
	List(ctx context.Context) ([]Result, error)
	// ListBySearch returns every Result sharing searchID, for progress
	// tracking and per-search no-improvement checks.
	ListBySearch(ctx context.Context, searchID string) ([]Result, error)
	// SaveTrades replaces the full trade list for one experiment — the
	// worker calls this once per completed backtest, not incrementally, so
	// a full replace (not append) keeps retries idempotent.
	SaveTrades(ctx context.Context, experimentID string, trades []Trade) error
	ListTrades(ctx context.Context, experimentID string) ([]Trade, error)
}
