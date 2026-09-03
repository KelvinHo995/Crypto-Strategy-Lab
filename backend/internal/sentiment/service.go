package sentiment

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/news"
)

var (
	ErrAnalyze   = errors.New("analyze sentiment")
	ErrFetchNews = errors.New("fetch news")
	ErrStore     = errors.New("store sentiment")
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

func (s *Service) IngestNews(ctx context.Context, items []news.NewsItem) ([]Observation, error) {
	observations := make([]Observation, 0, len(items))
	if len(items) == 0 {
		return observations, nil
	}
	if s == nil || s.analyzer == nil || s.repo == nil {
		return observations, fmt.Errorf("%w: service unavailable", ErrAnalyze)
	}

	uniqueItems := make([]news.NewsItem, 0, len(items))
	newsIDs := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item.ID = strings.TrimSpace(item.ID)
		if _, duplicate := seen[item.ID]; duplicate {
			continue
		}
		seen[item.ID] = struct{}{}
		uniqueItems = append(uniqueItems, item)
		newsIDs = append(newsIDs, item.ID)
	}

	existing, err := s.repo.ExistingNewsIDs(ctx, newsIDs)
	if err != nil {
		return observations, fmt.Errorf("%w: check existing news IDs: %w", ErrStore, err)
	}

	for _, item := range uniqueItems {
		if _, alreadyAnalyzed := existing[item.ID]; alreadyAnalyzed {
			continue
		}
		observation, err := s.AnalyzeAndStore(ctx, item.ID, item.Text, item.PublishedAt)
		if err != nil {
			return observations, fmt.Errorf("ingest news %q: %w", item.ID, err)
		}
		observations = append(observations, observation)
	}
	return observations, nil
}

func (s *Service) IngestFromProvider(ctx context.Context, provider news.NewsProvider, sinceTimestamp int64) ([]Observation, error) {
	if provider == nil {
		return nil, fmt.Errorf("%w: provider is required", ErrFetchNews)
	}
	items, fetchErr := provider.Fetch(ctx, sinceTimestamp)
	observations, ingestErr := s.IngestNews(ctx, items)
	if ingestErr != nil {
		return observations, ingestErr
	}
	if fetchErr != nil {
		return observations, fmt.Errorf("%w: %w", ErrFetchNews, fetchErr)
	}
	return observations, nil
}
