package sentiment

import (
	"container/list"
	"context"
	"sync"
	"time"
)

type ScoreFetcher interface {
	FetchSentiment(ctx context.Context, timestamp int64) (float64, error)
}

type scoreEntry struct {
	timestamp int64
	score     float64
	expiresAt time.Time
}

type scoreLoad struct {
	done  chan struct{}
	score float64
	err   error
}

// MemoizedLookup prevents repeated candidate evaluations over the same candle
// timestamps from issuing identical Postgres sentiment lookups. Errors are not
// cached, so a temporarily missing observation can recover immediately.
type MemoizedLookup struct {
	next       ScoreFetcher
	maxEntries int
	ttl        time.Duration

	mu       sync.Mutex
	items    map[int64]*list.Element
	lru      *list.List
	inflight map[int64]*scoreLoad
}

func NewMemoizedLookup(next ScoreFetcher, maxEntries int, ttl time.Duration) *MemoizedLookup {
	if maxEntries < 1 {
		maxEntries = 1
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &MemoizedLookup{next: next, maxEntries: maxEntries, ttl: ttl, items: make(map[int64]*list.Element), lru: list.New(), inflight: make(map[int64]*scoreLoad)}
}

func (l *MemoizedLookup) FetchSentiment(ctx context.Context, timestamp int64) (float64, error) {
	now := time.Now()
	l.mu.Lock()
	if element, ok := l.items[timestamp]; ok {
		entry := element.Value.(*scoreEntry)
		if now.Before(entry.expiresAt) {
			l.lru.MoveToFront(element)
			l.mu.Unlock()
			return entry.score, nil
		}
		l.removeElement(element)
	}
	if load, ok := l.inflight[timestamp]; ok {
		l.mu.Unlock()
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-load.done:
			return load.score, load.err
		}
	}
	load := &scoreLoad{done: make(chan struct{})}
	l.inflight[timestamp] = load
	l.mu.Unlock()

	score, err := l.next.FetchSentiment(ctx, timestamp)
	l.mu.Lock()
	load.score, load.err = score, err
	delete(l.inflight, timestamp)
	if err == nil {
		l.items[timestamp] = l.lru.PushFront(&scoreEntry{timestamp: timestamp, score: score, expiresAt: time.Now().Add(l.ttl)})
		for l.lru.Len() > l.maxEntries {
			l.removeElement(l.lru.Back())
		}
	}
	close(load.done)
	l.mu.Unlock()
	return score, err
}

func (l *MemoizedLookup) removeElement(element *list.Element) {
	if element == nil {
		return
	}
	entry := element.Value.(*scoreEntry)
	delete(l.items, entry.timestamp)
	l.lru.Remove(element)
}
