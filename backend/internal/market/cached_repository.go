package market

import (
	"container/list"
	"context"
	"sync"
	"time"
)

type candleCacheKey struct {
	symbol, timeframe string
	from, to          int64
	generation        uint64
}

type candleCacheEntry struct {
	key       candleCacheKey
	candles   []Candle
	expiresAt time.Time
}

type candleLoad struct {
	done    chan struct{}
	candles []Candle
	err     error
}

// CachedCandleRepository is a bounded process-local read-through cache. It is
// intentionally a decorator around CandleRepository so Postgres remains the
// source of truth and Redis is not required for a single backend instance.
type CachedCandleRepository struct {
	next       CandleRepository
	maxEntries int
	ttl        time.Duration

	mu         sync.Mutex
	items      map[candleCacheKey]*list.Element
	lru        *list.List
	inflight   map[candleCacheKey]*candleLoad
	generation uint64
}

func NewCachedCandleRepository(next CandleRepository, maxEntries int, ttl time.Duration) *CachedCandleRepository {
	if maxEntries < 1 {
		maxEntries = 1
	}
	if ttl <= 0 {
		ttl = time.Minute
	}
	return &CachedCandleRepository{next: next, maxEntries: maxEntries, ttl: ttl, items: make(map[candleCacheKey]*list.Element), lru: list.New(), inflight: make(map[candleCacheKey]*candleLoad)}
}

func (r *CachedCandleRepository) Upsert(ctx context.Context, candles []Candle) error {
	if err := r.next.Upsert(ctx, candles); err != nil {
		return err
	}
	r.clear()
	return nil
}

func (r *CachedCandleRepository) Range(ctx context.Context, symbol, timeframe string, from, to int64) ([]Candle, error) {
	now := time.Now()
	r.mu.Lock()
	key := candleCacheKey{symbol: symbol, timeframe: timeframe, from: from, to: to, generation: r.generation}
	if element, ok := r.items[key]; ok {
		entry := element.Value.(*candleCacheEntry)
		if now.Before(entry.expiresAt) {
			r.lru.MoveToFront(element)
			candles := cloneCandles(entry.candles)
			r.mu.Unlock()
			return candles, nil
		}
		r.removeElement(element)
	}
	if load, ok := r.inflight[key]; ok {
		r.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-load.done:
			return cloneCandles(load.candles), load.err
		}
	}
	load := &candleLoad{done: make(chan struct{})}
	r.inflight[key] = load
	r.mu.Unlock()

	candles, err := r.next.Range(ctx, symbol, timeframe, from, to)
	loaded := cloneCandles(candles)
	r.mu.Lock()
	load.candles, load.err = loaded, err
	delete(r.inflight, key)
	if err == nil {
		entry := &candleCacheEntry{key: key, candles: cloneCandles(loaded), expiresAt: time.Now().Add(r.ttl)}
		r.items[key] = r.lru.PushFront(entry)
		for r.lru.Len() > r.maxEntries {
			r.removeElement(r.lru.Back())
		}
	}
	close(load.done)
	r.mu.Unlock()
	return cloneCandles(loaded), err
}

func (r *CachedCandleRepository) clear() {
	r.mu.Lock()
	r.generation++
	r.items = make(map[candleCacheKey]*list.Element)
	r.lru.Init()
	r.mu.Unlock()
}

func (r *CachedCandleRepository) removeElement(element *list.Element) {
	if element == nil {
		return
	}
	entry := element.Value.(*candleCacheEntry)
	delete(r.items, entry.key)
	r.lru.Remove(element)
}

func cloneCandles(source []Candle) []Candle {
	if source == nil {
		return nil
	}
	return append([]Candle(nil), source...)
}
