package market_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

type countingCandleRepository struct {
	mu    sync.Mutex
	loads int
}

func (r *countingCandleRepository) Upsert(context.Context, []market.Candle) error { return nil }
func (r *countingCandleRepository) Range(context.Context, string, string, int64, int64) ([]market.Candle, error) {
	time.Sleep(10 * time.Millisecond)
	r.mu.Lock()
	r.loads++
	r.mu.Unlock()
	return []market.Candle{{Symbol: "BTCUSDT", Close: 100}}, nil
}

func TestCachedCandleRepositoryCoalescesAndClones(t *testing.T) {
	next := &countingCandleRepository{}
	repo := market.NewCachedCandleRepository(next, 4, time.Minute)
	var wg sync.WaitGroup
	results := make([][]market.Candle, 8)
	for i := range results {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index], _ = repo.Range(context.Background(), "BTCUSDT", "1h", 1, 2)
		}(i)
	}
	wg.Wait()
	if next.loads != 1 {
		t.Fatalf("underlying loads = %d, want 1", next.loads)
	}
	results[0][0].Close = 1
	again, err := repo.Range(context.Background(), "BTCUSDT", "1h", 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if again[0].Close != 100 {
		t.Fatal("caller mutation leaked into cached candle data")
	}
}
