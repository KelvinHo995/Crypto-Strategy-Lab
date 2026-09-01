package experiment

import (
	"context"
	"sync"
	"time"
)

type resultListLoad struct {
	generation uint64
	done       chan struct{}
	results    []Result
	err        error
}

// CachedRepository caches the bounded leaderboard snapshot. Writes always go
// to the source repository first and invalidate the snapshot immediately.
type CachedRepository struct {
	next Repository
	ttl  time.Duration

	mu         sync.Mutex
	results    []Result
	expiresAt  time.Time
	inflight   map[uint64]*resultListLoad
	generation uint64
}

func NewCachedRepository(next Repository, ttl time.Duration) *CachedRepository {
	if ttl <= 0 {
		ttl = 3 * time.Second
	}
	return &CachedRepository{next: next, ttl: ttl, inflight: make(map[uint64]*resultListLoad)}
}

func (r *CachedRepository) Save(ctx context.Context, result Result) error {
	if err := r.next.Save(ctx, result); err != nil {
		return err
	}
	r.invalidate()
	return nil
}

func (r *CachedRepository) Get(ctx context.Context, id string) (Result, error) {
	return r.next.Get(ctx, id)
}

// ListBySearch isn't cached — it's a per-search progress/no-improvement
// lookup, not the hot bounded-leaderboard read List() exists to protect.
func (r *CachedRepository) ListBySearch(ctx context.Context, searchID string) ([]Result, error) {
	return r.next.ListBySearch(ctx, searchID)
}

func (r *CachedRepository) List(ctx context.Context) ([]Result, error) {
	r.mu.Lock()
	generation := r.generation
	if r.results != nil && time.Now().Before(r.expiresAt) {
		results := cloneResults(r.results)
		r.mu.Unlock()
		return results, nil
	}
	if load := r.inflight[generation]; load != nil {
		r.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-load.done:
			return cloneResults(load.results), load.err
		}
	}
	load := &resultListLoad{done: make(chan struct{}), generation: generation}
	r.inflight[generation] = load
	r.mu.Unlock()

	results, err := r.next.List(ctx)
	r.mu.Lock()
	load.results, load.err = cloneResults(results), err
	if err == nil && generation == r.generation {
		r.results = cloneResults(results)
		r.expiresAt = time.Now().Add(r.ttl)
	}
	delete(r.inflight, generation)
	close(load.done)
	r.mu.Unlock()
	return cloneResults(results), err
}

func (r *CachedRepository) MarkStaleRunningFailed(ctx context.Context, olderThan time.Duration) ([]Result, error) {
	results, err := r.next.MarkStaleRunningFailed(ctx, olderThan)
	if err == nil && len(results) > 0 {
		r.invalidate()
	}
	return results, err
}

func (r *CachedRepository) invalidate() {
	r.mu.Lock()
	r.generation++
	r.results = nil
	r.expiresAt = time.Time{}
	r.mu.Unlock()
}

func (r *CachedRepository) Invalidate() { r.invalidate() }

func cloneResults(source []Result) []Result {
	if source == nil {
		return nil
	}
	results := make([]Result, len(source))
	for index, result := range source {
		result.Strategies = append([]string(nil), result.Strategies...)
		result.Params = cloneMap(result.Params)
		result.StrategyVersions = cloneStringMap(result.StrategyVersions)
		results[index] = result
	}
	return results
}
