package sentiment

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("sentiment observation not found")

type Observation struct {
	NewsID       string  `json:"newsId"`
	PublishedAt  int64   `json:"publishedAt"`
	Sentiment    string  `json:"sentiment"`
	Score        float64 `json:"score"`
	ModelName    string  `json:"modelName"`
	ModelVersion string  `json:"modelVersion"`
	AnalyzedAt   int64   `json:"analyzedAt"`
}

type Repository interface {
	Save(ctx context.Context, observation Observation) error
	// ListSince returns every observation with PublishedAt >= earliestPublishedAt,
	// ordered ascending by PublishedAt, with no upper bound — callers needing a
	// point-in-time lookup (see TimeLookup) filter the result themselves, which
	// lets them cache it across many lookups instead of querying per call.
	ListSince(ctx context.Context, earliestPublishedAt int64) ([]Observation, error)
}
