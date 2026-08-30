package httpx_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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

func (f *fakeSentimentService) AnalyzeAndStore(_ context.Context, newsID, text string, publishedAt int64) (sentiment.Observation, error) {
	f.newsID = newsID
	f.text = text
	f.publishedAt = publishedAt
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
