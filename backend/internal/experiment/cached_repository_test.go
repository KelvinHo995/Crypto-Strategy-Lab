package experiment_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
)

type countingExperimentRepository struct {
	mu    sync.Mutex
	lists int
}

func (r *countingExperimentRepository) Save(context.Context, experiment.Result) error { return nil }
func (r *countingExperimentRepository) Get(context.Context, string) (experiment.Result, error) {
	return experiment.Result{}, experiment.ErrNotFound
}
func (r *countingExperimentRepository) List(context.Context) ([]experiment.Result, error) {
	time.Sleep(10 * time.Millisecond)
	r.mu.Lock()
	r.lists++
	r.mu.Unlock()
	return []experiment.Result{{ID: "one", Strategies: []string{"MA"}}}, nil
}
func (r *countingExperimentRepository) MarkStaleRunningFailed(context.Context, time.Duration) ([]experiment.Result, error) {
	return nil, nil
}

func TestCachedExperimentRepositoryCoalescesAndInvalidates(t *testing.T) {
	next := &countingExperimentRepository{}
	repo := experiment.NewCachedRepository(next, time.Minute)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = repo.List(context.Background())
		}()
	}
	wg.Wait()
	if next.lists != 1 {
		t.Fatalf("list calls = %d, want 1", next.lists)
	}
	if err := repo.Save(context.Background(), experiment.Result{ID: "two"}); err != nil {
		t.Fatal(err)
	}
	_, _ = repo.List(context.Background())
	if next.lists != 2 {
		t.Fatalf("list calls after write = %d, want 2", next.lists)
	}
}
