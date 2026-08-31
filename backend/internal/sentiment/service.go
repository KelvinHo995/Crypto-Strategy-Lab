package sentiment

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrAnalyze = errors.New("analyze sentiment")
	ErrStore   = errors.New("store sentiment")
)

type Analyzer interface {
	Analyze(ctx context.Context, newsID, text string) (Result, error)
}

type Service struct {
	analyzer Analyzer
	repo     Repository
}

func NewService(analyzer Analyzer, repo Repository) *Service {
	return &Service{analyzer: analyzer, repo: repo}
}

func (s *Service) AnalyzeAndStore(ctx context.Context, newsID, text string, publishedAt int64) (Observation, error) {
	if s == nil || s.analyzer == nil || s.repo == nil {
		return Observation{}, fmt.Errorf("%w: service unavailable", ErrAnalyze)
	}
	newsID = strings.TrimSpace(newsID)
	if newsID == "" || strings.TrimSpace(text) == "" || publishedAt <= 0 {
		return Observation{}, fmt.Errorf("%w: news id, text, and publishedAt are required", ErrAnalyze)
	}

	result, err := s.analyzer.Analyze(ctx, newsID, text)
	if err != nil {
		return Observation{}, fmt.Errorf("%w: %v", ErrAnalyze, err)
	}
	observation := Observation{
		NewsID:       result.NewsID,
		PublishedAt:  publishedAt,
		Sentiment:    result.Sentiment,
		Score:        result.Score,
		ModelName:    result.Model.Name,
		ModelVersion: result.Model.Version,
		AnalyzedAt:   result.CreatedAt * 1000,
	}
	if err := s.repo.Save(ctx, observation); err != nil {
		return Observation{}, fmt.Errorf("%w: %v", ErrStore, err)
	}
	return observation, nil
}
