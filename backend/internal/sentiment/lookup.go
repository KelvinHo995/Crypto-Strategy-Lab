package sentiment

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

const DefaultMaxAge = 24 * time.Hour

// cacheTTL bounds how stale the cache can get before a lookup forces a
// refresh — a background safety net for writers that don't call Invalidate,
// not the primary correctness mechanism (that's Invalidate itself). One
// backtest run needs at most one real fetch regardless of this value, since
// candles are processed in ascending time order within well under a minute.
const cacheTTL = time.Minute

type TimeLookup struct {
	repo   Repository
	maxAge time.Duration

	mu             sync.Mutex
	cached         []Observation // sorted ascending by PublishedAt
	cachedAt       time.Time
	cachedEarliest int64 // the earliestPublishedAt bound the current cache covers
}

func NewTimeLookup(repo Repository, maxAge time.Duration) *TimeLookup {
	if maxAge <= 0 || maxAge > DefaultMaxAge {
		maxAge = DefaultMaxAge
	}
	return &TimeLookup{repo: repo, maxAge: maxAge}
}

// Invalidate forces the next lookup to refetch from the repository instead
// of trusting whatever's cached. Call this after writing an observation the
// cache wouldn't otherwise know about — the TTL alone only catches it after
// up to a minute's delay.
func (l *TimeLookup) Invalidate() {
	l.mu.Lock()
	l.cached = nil
	l.mu.Unlock()
}

// FetchSentiment returns the directional 0..1 score expected by SentimentStrategy.
func (l *TimeLookup) FetchSentiment(ctx context.Context, timestamp int64) (float64, error) {
	observation, err := l.fetchObservation(ctx, timestamp)
	if err != nil {
		return 0, err
	}
	return directionalScore(observation)
}

// FetchSentimentWithModel preserves the identity of the exact persisted
// observation selected for a strategy lookup.
func (l *TimeLookup) FetchSentimentWithModel(ctx context.Context, timestamp int64) (float64, string, string, error) {
	observation, err := l.fetchObservation(ctx, timestamp)
	if err != nil {
		return 0, "", "", err
	}
	score, err := directionalScore(observation)
	if err != nil {
		return 0, "", "", err
	}
	modelName := strings.TrimSpace(observation.ModelName)
	modelVersion := strings.TrimSpace(observation.ModelVersion)
	if modelName == "" || modelVersion == "" {
		return 0, "", "", errors.New("sentiment observation has no model identity")
	}
	return score, modelName, modelVersion, nil
}

func (l *TimeLookup) fetchObservation(ctx context.Context, timestamp int64) (Observation, error) {
	if l == nil || l.repo == nil {
		return Observation{}, errors.New("sentiment repository unavailable")
	}
	if timestamp <= 0 {
		return Observation{}, errors.New("sentiment timestamp must be positive")
	}
	earliest := timestamp - l.maxAge.Milliseconds()

	observation, err := l.latestAtOrBefore(ctx, timestamp, earliest)
	if err != nil {
		return Observation{}, err
	}
	return observation, nil
}

func directionalScore(observation Observation) (float64, error) {
	switch observation.Sentiment {
	case "POSITIVE":
		return observation.Score, nil
	case "NEGATIVE":
		return 1 - observation.Score, nil
	case "NEUTRAL":
		return 0.5, nil
	default:
		return 0, fmt.Errorf("unknown sentiment label %q", observation.Sentiment)
	}
}

func (l *TimeLookup) latestAtOrBefore(ctx context.Context, timestamp, earliest int64) (Observation, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.cached == nil || earliest < l.cachedEarliest || time.Since(l.cachedAt) > cacheTTL {
		observations, err := l.repo.ListSince(ctx, earliest)
		if err != nil {
			return Observation{}, err
		}
		sort.Slice(observations, func(i, j int) bool { return observations[i].PublishedAt < observations[j].PublishedAt })
		l.cached = observations
		l.cachedAt = time.Now()
		l.cachedEarliest = earliest
	}

	idx := sort.Search(len(l.cached), func(i int) bool { return l.cached[i].PublishedAt > timestamp })
	if idx == 0 {
		return Observation{}, ErrNotFound
	}
	best := l.cached[idx-1]
	if best.PublishedAt < earliest {
		return Observation{}, ErrNotFound
	}
	return best, nil
}
