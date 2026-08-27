package experiment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
)

func TestInMemoryQueueHonorsCancellation(t *testing.T) {
	q := experiment.NewInMemoryQueue(1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := q.Dequeue(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Dequeue error = %v, want context.Canceled", err)
	}
}

func TestInMemoryQueueRoundTrip(t *testing.T) {
	q := experiment.NewInMemoryQueue(1)
	want := experiment.BacktestJob{ID: "job-1"}
	if err := q.Enqueue(context.Background(), want); err != nil {
		t.Fatal(err)
	}
	got, err := q.Dequeue(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != want.ID {
		t.Fatalf("job ID = %q, want %q", got.ID, want.ID)
	}
}
