package experiment_test

import (
	"context"
	"sync"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
)

type memRepo struct {
	mu      sync.Mutex
	results map[string]experiment.Result
	trades  map[string][]experiment.Trade
}

func newMemRepo(seed ...experiment.Result) *memRepo {
	r := &memRepo{results: make(map[string]experiment.Result), trades: make(map[string][]experiment.Trade)}
	for _, s := range seed {
		r.results[s.ID] = s
	}
	return r
}

func (r *memRepo) Save(_ context.Context, res experiment.Result) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.results[res.ID] = res
	return nil
}
func (r *memRepo) Get(_ context.Context, id string) (experiment.Result, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	res, ok := r.results[id]
	if !ok {
		return experiment.Result{}, experiment.ErrNotFound
	}
	return res, nil
}
func (r *memRepo) List(context.Context) ([]experiment.Result, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]experiment.Result, 0, len(r.results))
	for _, res := range r.results {
		out = append(out, res)
	}
	return out, nil
}
func (r *memRepo) ListBySearch(_ context.Context, searchID string) ([]experiment.Result, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []experiment.Result
	for _, res := range r.results {
		if res.SearchID == searchID {
			out = append(out, res)
		}
	}
	return out, nil
}

func (r *memRepo) SaveTrades(_ context.Context, experimentID string, trades []experiment.Trade) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.trades[experimentID] = trades
	return nil
}

func (r *memRepo) ListTrades(_ context.Context, experimentID string) ([]experiment.Trade, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.trades[experimentID], nil
}
