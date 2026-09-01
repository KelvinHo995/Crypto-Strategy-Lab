package sentiment_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/sentiment"
)

type countingScoreFetcher struct {
	mu    sync.Mutex
	calls int
	err   error
}

func (f *countingScoreFetcher) FetchSentiment(context.Context, int64) (float64, error) {
	time.Sleep(10 * time.Millisecond)
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()
	return 0.8, f.err
}

func TestMemoizedLookupCoalescesSuccessButNotErrors(t *testing.T) {
	fetcher := &countingScoreFetcher{}
	lookup := sentiment.NewMemoizedLookup(fetcher, 8, time.Minute)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if score, err := lookup.FetchSentiment(context.Background(), 123); err != nil || score != 0.8 {
				t.Errorf("score=%v err=%v", score, err)
			}
		}()
	}
	wg.Wait()
	if fetcher.calls != 1 {
		t.Fatalf("calls=%d want 1", fetcher.calls)
	}

	failing := &countingScoreFetcher{err: errors.New("temporary")}
	errorLookup := sentiment.NewMemoizedLookup(failing, 8, time.Minute)
	_, _ = errorLookup.FetchSentiment(context.Background(), 456)
	_, _ = errorLookup.FetchSentiment(context.Background(), 456)
	if failing.calls != 2 {
		t.Fatalf("error calls=%d want 2", failing.calls)
	}
}
