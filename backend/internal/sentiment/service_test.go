package sentiment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/sentiment"
)

type fakeAnalyzer struct {
	result sentiment.Result
	err    error
}

func (f *fakeAnalyzer) Analyze(context.Context, string, string) (sentiment.Result, error) {
	return f.result, f.err
}

type fakeRepository struct {
	saved       sentiment.Observation
	observation sentiment.Observation
	err         error
	latest      int64
	earliest    int64
}

func (f *fakeRepository) Save(_ context.Context, observation sentiment.Observation) error {
	f.saved = observation
	return f.err
}

func (f *fakeRepository) LatestAtOrBefore(_ context.Context, timestamp, earliestTimestamp int64) (sentiment.Observation, error) {
	f.latest = timestamp
	f.earliest = earliestTimestamp
	return f.observation, f.err
}

func TestServiceAnalyzeAndStore(t *testing.T) {
	var result sentiment.Result
	result.NewsID = "news-1"
	result.Sentiment = "POSITIVE"
	result.Score = 0.9
	result.Model.Name = "crypto-lexicon"
	result.Model.Version = "v1"
	result.CreatedAt = 123

	repo := &fakeRepository{}
	service := sentiment.NewService(&fakeAnalyzer{result: result}, repo)
	observation, err := service.AnalyzeAndStore(context.Background(), "news-1", "bullish rally", 456_000)
	if err != nil {
		t.Fatal(err)
	}
	if observation != repo.saved {
		t.Fatalf("returned observation = %+v, saved = %+v", observation, repo.saved)
	}
	if observation.PublishedAt != 456_000 || observation.AnalyzedAt != 123_000 {
		t.Fatalf("timestamps = published %d analyzed %d", observation.PublishedAt, observation.AnalyzedAt)
	}
	if observation.ModelName != "crypto-lexicon" || observation.ModelVersion != "v1" {
		t.Fatalf("model metadata = %+v", observation)
	}
}

func TestServiceReportsAnalysisAndStorageErrors(t *testing.T) {
	analyzeFailure := sentiment.NewService(&fakeAnalyzer{err: errors.New("service down")}, &fakeRepository{})
	if _, err := analyzeFailure.AnalyzeAndStore(context.Background(), "news-1", "text", 1); !errors.Is(err, sentiment.ErrAnalyze) {
		t.Fatalf("analysis error = %v", err)
	}

	var result sentiment.Result
	result.NewsID = "news-1"
	storeFailure := sentiment.NewService(&fakeAnalyzer{result: result}, &fakeRepository{err: errors.New("database down")})
	if _, err := storeFailure.AnalyzeAndStore(context.Background(), "news-1", "text", 1); !errors.Is(err, sentiment.ErrStore) {
		t.Fatalf("storage error = %v", err)
	}
}
