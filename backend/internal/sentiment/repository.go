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
	LatestAtOrBefore(ctx context.Context, timestamp, earliestTimestamp int64) (Observation, error)
}
