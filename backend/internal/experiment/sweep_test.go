package experiment_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
)

type memRepo struct {
	mu      sync.Mutex
	results map[string]experiment.Result
}

func newMemRepo(seed ...experiment.Result) *memRepo {
	r := &memRepo{results: make(map[string]experiment.Result)}
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
func (r *memRepo) MarkStaleRunningFailed(_ context.Context, olderThan time.Duration) ([]experiment.Result, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cutoff := time.Now().Add(-olderThan).UnixMilli()
	var stale []experiment.Result
	for id, res := range r.results {
		if res.Status != "RUNNING" || res.UpdatedAt >= cutoff {
			continue
		}
		res.Status = "FAILED"
		r.results[id] = res
		stale = append(stale, res)
	}
	return stale, nil
}

func TestSweep_MarksStaleRunningAsFailed(t *testing.T) {
	old := time.Now().Add(-20 * time.Minute).UnixMilli()
	repo := newMemRepo(experiment.Result{ID: "stuck", Status: "RUNNING", UpdatedAt: old})

	var notified experiment.Result
	experiment.NewSweeper().Sweep(context.Background(), repo, 15*time.Minute, func(r experiment.Result) {
		notified = r
	})

	got, err := repo.Get(context.Background(), "stuck")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "FAILED" {
		t.Fatalf("Status = %q, want FAILED", got.Status)
	}
	if notified.ID != "stuck" {
		t.Fatal("notify was not called with the swept result")
	}
}

func TestSweep_LeavesRecentAndNonRunningAlone(t *testing.T) {
	recent := time.Now().UnixMilli()
	old := time.Now().Add(-20 * time.Minute).UnixMilli()
	repo := newMemRepo(
		experiment.Result{ID: "fresh-running", Status: "RUNNING", UpdatedAt: recent},
		experiment.Result{ID: "old-completed", Status: "COMPLETED", UpdatedAt: old},
		experiment.Result{ID: "old-pending", Status: "PENDING", UpdatedAt: old},
	)

	calls := 0
	experiment.NewSweeper().Sweep(context.Background(), repo, 15*time.Minute, func(experiment.Result) { calls++ })

	if calls != 0 {
		t.Fatalf("notify called %d times, want 0", calls)
	}
	for _, id := range []string{"fresh-running", "old-completed", "old-pending"} {
		got, err := repo.Get(context.Background(), id)
		if err != nil {
			t.Fatal(err)
		}
		if got.Status == "FAILED" {
			t.Fatalf("%s was marked FAILED, want untouched", id)
		}
	}
}

func TestSweeperRun_StopsOnContextCancel(t *testing.T) {
	repo := newMemRepo()
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		experiment.NewSweeper().Run(ctx, repo, time.Minute, 10*time.Millisecond, nil)
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run did not return after context cancellation")
	}
}
