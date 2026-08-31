package sentiment

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const DefaultMaxAge = 24 * time.Hour

type TimeLookup struct {
	repo   Repository
	maxAge time.Duration
}

func NewTimeLookup(repo Repository, maxAge time.Duration) *TimeLookup {
	if maxAge <= 0 || maxAge > DefaultMaxAge {
		maxAge = DefaultMaxAge
	}
	return &TimeLookup{repo: repo, maxAge: maxAge}
}

// FetchSentiment returns the directional 0..1 score expected by SentimentStrategy.
func (l *TimeLookup) FetchSentiment(ctx context.Context, timestamp int64) (float64, error) {
	if l == nil || l.repo == nil {
		return 0, errors.New("sentiment repository unavailable")
	}
	if timestamp <= 0 {
		return 0, errors.New("sentiment timestamp must be positive")
	}
	earliest := timestamp - l.maxAge.Milliseconds()
	observation, err := l.repo.LatestAtOrBefore(ctx, timestamp, earliest)
	if err != nil {
		return 0, err
	}

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
