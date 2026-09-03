package sentiment

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("sentiment observation not found")

type Observation struct {
	NewsID       string  `json:"newsId"`
	Title        string  `json:"title,omitempty"`
	Source       string  `json:"source,omitempty"`
	URL          string  `json:"url,omitempty"`
	PublishedAt  int64   `json:"publishedAt"`
	Sentiment    string  `json:"sentiment"`
	Score        float64 `json:"score"`
	ModelName    string  `json:"modelName"`
	ModelVersion string  `json:"modelVersion"`
	AnalyzedAt   int64   `json:"analyzedAt"`
}

type Repository interface {
	Save(ctx context.Context, observation Observation) error
	// ExistingNewsIDs returns the subset of news IDs already persisted.
	ExistingNewsIDs(ctx context.Context, newsIDs []string) (map[string]struct{}, error)
	// ListSince returns every observation with PublishedAt >= earliestPublishedAt,
	// ordered ascending by PublishedAt, with no upper bound — callers needing a
	// point-in-time lookup (see TimeLookup) filter the result themselves, which
	// lets them cache it across many lookups instead of querying per call.
	ListSince(ctx context.Context, earliestPublishedAt int64) ([]Observation, error)
}
