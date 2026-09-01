package sentiment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/news"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/sentiment"
)

type recordingAnalyzer struct {
	calls  []string
	texts  []string
	failID string
}

func (a *recordingAnalyzer) Analyze(_ context.Context, newsID, text string) (sentiment.Result, error) {
	a.calls = append(a.calls, newsID)
	a.texts = append(a.texts, text)
	if newsID == a.failID {
		return sentiment.Result{}, errors.New("analyzer unavailable")
	}
	var result sentiment.Result
	result.NewsID = newsID
	result.Sentiment = "POSITIVE"
	result.Score = 0.9
	result.Model.Name = "test-model"
	result.Model.Version = "v1"
	result.CreatedAt = 1
	return result, nil
}

type recordingRepository struct {
	observations map[string]sentiment.Observation
	saves        int
}

func (r *recordingRepository) Save(_ context.Context, observation sentiment.Observation) error {
	if r.observations == nil {
		r.observations = make(map[string]sentiment.Observation)
	}
	r.observations[observation.NewsID] = observation
	r.saves++
	return nil
}

func (r *recordingRepository) ListSince(context.Context, int64) ([]sentiment.Observation, error) {
	return nil, nil
}

type fakeNewsProvider struct {
	items []news.NewsItem
	err   error
	since int64
}

func (p *fakeNewsProvider) Fetch(_ context.Context, sinceTimestamp int64) ([]news.NewsItem, error) {
	p.since = sinceTimestamp
	return p.items, p.err
}

func testNewsItem(id string) news.NewsItem {
	return news.NewsItem{
		ID:          id,
		Title:       "Bitcoin rally continues",
		Text:        "Bitcoin posts bullish gains",
		Source:      "Test Source",
		URL:         "https://example.com/" + id,
		PublishedAt: 1_000,
	}
}

func TestIngestOneArticleFromProvider(t *testing.T) {
	analyzer := &recordingAnalyzer{}
	repo := &recordingRepository{}
	provider := &fakeNewsProvider{items: []news.NewsItem{testNewsItem("news-1")}}
	service := sentiment.NewService(analyzer, repo)

	observations, err := service.IngestFromProvider(context.Background(), provider, 500)
	if err != nil {
		t.Fatal(err)
	}
	if provider.since != 500 || len(observations) != 1 || len(repo.observations) != 1 {
		t.Fatalf("since=%d observations=%d stored=%d", provider.since, len(observations), len(repo.observations))
	}
	if analyzer.calls[0] != "news-1" || analyzer.texts[0] != "Bitcoin posts bullish gains" {
		t.Fatalf("analyzer input = %q %q", analyzer.calls[0], analyzer.texts[0])
	}
}

func TestIngestMultipleArticles(t *testing.T) {
	analyzer := &recordingAnalyzer{}
	repo := &recordingRepository{}
	service := sentiment.NewService(analyzer, repo)

	observations, err := service.IngestNews(context.Background(), []news.NewsItem{
		testNewsItem("news-1"),
		testNewsItem("news-2"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(observations) != 2 || len(analyzer.calls) != 2 || len(repo.observations) != 2 {
		t.Fatalf("observations=%d calls=%d stored=%d", len(observations), len(analyzer.calls), len(repo.observations))
	}
}

func TestIngestDuplicateArticleIsIdempotent(t *testing.T) {
	analyzer := &recordingAnalyzer{}
	repo := &recordingRepository{}
	service := sentiment.NewService(analyzer, repo)
	article := testNewsItem("news-1")

	observations, err := service.IngestNews(context.Background(), []news.NewsItem{article, article})
	if err != nil {
		t.Fatal(err)
	}
	if len(observations) != 1 || len(analyzer.calls) != 1 || len(repo.observations) != 1 {
		t.Fatalf("batch observations=%d calls=%d stored=%d", len(observations), len(analyzer.calls), len(repo.observations))
	}

	if _, err := service.IngestNews(context.Background(), []news.NewsItem{article}); err != nil {
		t.Fatal(err)
	}
	if len(repo.observations) != 1 || repo.saves != 2 {
		t.Fatalf("stored=%d saves=%d, repeated ingestion should upsert one logical row", len(repo.observations), repo.saves)
	}
}

func TestIngestStopsOnAnalyzerFailure(t *testing.T) {
	analyzer := &recordingAnalyzer{failID: "news-bad"}
	repo := &recordingRepository{}
	service := sentiment.NewService(analyzer, repo)

	_, err := service.IngestNews(context.Background(), []news.NewsItem{testNewsItem("news-bad")})
	if !errors.Is(err, sentiment.ErrAnalyze) {
		t.Fatalf("error = %v, want ErrAnalyze", err)
	}
	if len(repo.observations) != 0 {
		t.Fatalf("stored=%d, failed analysis must not be persisted", len(repo.observations))
	}
}

func TestIngestReportsProviderFailure(t *testing.T) {
	analyzer := &recordingAnalyzer{}
	repo := &recordingRepository{}
	provider := &fakeNewsProvider{err: errors.New("source unavailable")}
	service := sentiment.NewService(analyzer, repo)

	_, err := service.IngestFromProvider(context.Background(), provider, 500)
	if !errors.Is(err, sentiment.ErrFetchNews) {
		t.Fatalf("error = %v, want ErrFetchNews", err)
	}
	if len(analyzer.calls) != 0 || repo.saves != 0 {
		t.Fatal("provider failure should not analyze or persist articles")
	}
}
