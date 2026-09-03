package httpx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/httpx"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/sentiment"
)

type fakeSentimentService struct {
	observation sentiment.Observation
	err         error
	newsID      string
	text        string
	publishedAt int64
}

func (f *fakeSentimentService) AnalyzeAndStore(_ context.Context, input sentiment.AnalyzeInput) (sentiment.Observation, error) {
	f.newsID = input.NewsID
	f.text = input.Text
	f.publishedAt = input.PublishedAt
	return f.observation, f.err
}

func TestAnalyzeSentimentStoresObservation(t *testing.T) {
	service := &fakeSentimentService{observation: sentiment.Observation{NewsID: "news-1", Sentiment: "POSITIVE", Score: 0.9}}
	router := httpx.NewRouterWithContext(context.Background(), newTestRegistry(), newFakeRepo(), httpx.Dependencies{Sentiment: service})
	defer router.Close()

	req := httptest.NewRequest(http.MethodPost, "/sentiment/analyze", bytes.NewBufferString(
		`{"newsId":" news-1 ","text":"bullish rally","publishedAt":123000}`,
	))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", w.Code)
	}
	if service.newsID != "news-1" || service.text != "bullish rally" || service.publishedAt != 123_000 {
		t.Fatalf("service input = %q %q %d", service.newsID, service.text, service.publishedAt)
	}
}

func TestAnalyzeSentimentRejectsInvalidInput(t *testing.T) {
	router := httpx.NewRouterWithContext(context.Background(), newTestRegistry(), newFakeRepo(), httpx.Dependencies{Sentiment: &fakeSentimentService{}})
	defer router.Close()
	req := httptest.NewRequest(http.MethodPost, "/sentiment/analyze", bytes.NewBufferString(
		`{"newsId":"","text":"","publishedAt":0}`,
	))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestAnalyzeSentimentMapsDependencyErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "analysis", err: sentiment.ErrAnalyze, want: http.StatusBadGateway},
		{name: "storage", err: sentiment.ErrStore, want: http.StatusInternalServerError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &fakeSentimentService{err: test.err}
			router := httpx.NewRouterWithContext(context.Background(), newTestRegistry(), newFakeRepo(), httpx.Dependencies{Sentiment: service})
			defer router.Close()
			req := httptest.NewRequest(http.MethodPost, "/sentiment/analyze", bytes.NewBufferString(
				`{"newsId":"news-1","text":"text","publishedAt":1}`,
			))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != test.want {
				t.Fatalf("status = %d, want %d", w.Code, test.want)
			}
		})
	}
}

type fakeSentimentReader struct {
	observations []sentiment.Observation
	err          error
	sinceCalled  int64
}

func (f *fakeSentimentReader) ListSince(_ context.Context, earliestPublishedAt int64) ([]sentiment.Observation, error) {
	f.sinceCalled = earliestPublishedAt
	return f.observations, f.err
}

func TestListSentimentObservations_ReturnsRecentFirst(t *testing.T) {
	reader := &fakeSentimentReader{observations: []sentiment.Observation{
		{NewsID: "old", Title: "Older article", PublishedAt: 1000},
		{NewsID: "new", Title: "Newer article", PublishedAt: 3000},
		{NewsID: "mid", Title: "Middle article", PublishedAt: 2000},
	}}
	router := httpx.NewRouterWithContext(context.Background(), newTestRegistry(), newFakeRepo(), httpx.Dependencies{SentimentReader: reader})
	defer router.Close()

	req := httptest.NewRequest(http.MethodGet, "/sentiment/observations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var got []sentiment.Observation
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0].NewsID != "new" || got[1].NewsID != "mid" || got[2].NewsID != "old" {
		t.Fatalf("order = %+v, want newest first", got)
	}
}

func TestListSentimentObservations_DefaultsSinceWindow(t *testing.T) {
	reader := &fakeSentimentReader{}
	router := httpx.NewRouterWithContext(context.Background(), newTestRegistry(), newFakeRepo(), httpx.Dependencies{SentimentReader: reader})
	defer router.Close()

	req := httptest.NewRequest(http.MethodGet, "/sentiment/observations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	dayAgo := time.Now().Add(-25 * time.Hour).UnixMilli()
	if reader.sinceCalled < dayAgo {
		t.Fatalf("since = %d, want roughly 24h ago (%d)", reader.sinceCalled, dayAgo)
	}
}

func TestListSentimentObservations_AcceptsSinceParam(t *testing.T) {
	reader := &fakeSentimentReader{}
	router := httpx.NewRouterWithContext(context.Background(), newTestRegistry(), newFakeRepo(), httpx.Dependencies{SentimentReader: reader})
	defer router.Close()

	req := httptest.NewRequest(http.MethodGet, "/sentiment/observations?since=500", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if reader.sinceCalled != 500 {
		t.Fatalf("since = %d, want 500", reader.sinceCalled)
	}
}
