package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/sentiment"
)

type SentimentService interface {
	AnalyzeAndStore(ctx context.Context, input sentiment.AnalyzeInput) (sentiment.Observation, error)
}

// SentimentReader is the narrow read-side capability the news feed needs —
// separate from SentimentService (write-side, single-article analyze) since
// a handler should only depend on what it actually calls.
type SentimentReader interface {
	ListSince(ctx context.Context, earliestPublishedAt int64) ([]sentiment.Observation, error)
}

type AnalyzeSentimentRequest struct {
	NewsID      string `json:"newsId"`
	Title       string `json:"title,omitempty"`
	Text        string `json:"text"`
	Source      string `json:"source,omitempty"`
	URL         string `json:"url,omitempty"`
	PublishedAt int64  `json:"publishedAt"`
}

func analyzeSentiment(service SentimentService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if service == nil {
			http.Error(w, "sentiment unavailable", http.StatusServiceUnavailable)
			return
		}

		var req AnalyzeSentimentRequest
		if err := decodeJSON(w, r, &req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		req.NewsID = strings.TrimSpace(req.NewsID)
		if req.NewsID == "" || strings.TrimSpace(req.Text) == "" || req.PublishedAt <= 0 {
			http.Error(w, "newsId, text, and publishedAt are required", http.StatusBadRequest)
			return
		}

		observation, err := service.AnalyzeAndStore(r.Context(), sentiment.AnalyzeInput{
			NewsID: req.NewsID, Title: req.Title, Text: req.Text,
			Source: req.Source, URL: req.URL, PublishedAt: req.PublishedAt,
		})
		if errors.Is(err, sentiment.ErrStore) {
			http.Error(w, "store sentiment", http.StatusInternalServerError)
			return
		}
		if err != nil {
			http.Error(w, "analyze sentiment", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(observation)
	}
}

// listSentimentObservations powers the real news feed — recent analyzed
// articles, most-recent first. Defaults to the last 24h; ?since=<unix-ms>
// widens or narrows the window.
func listSentimentObservations(reader SentimentReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if reader == nil {
			http.Error(w, "sentiment unavailable", http.StatusServiceUnavailable)
			return
		}
		since := time.Now().Add(-24 * time.Hour).UnixMilli()
		if raw := r.URL.Query().Get("since"); raw != "" {
			parsed, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				http.Error(w, "since must be a unix millisecond timestamp", http.StatusBadRequest)
				return
			}
			since = parsed
		}

		observations, err := reader.ListSince(r.Context(), since)
		if err != nil {
			http.Error(w, "load sentiment observations", http.StatusInternalServerError)
			return
		}
		if observations == nil {
			observations = []sentiment.Observation{}
		}
		sort.Slice(observations, func(i, j int) bool { return observations[i].PublishedAt > observations[j].PublishedAt })

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(observations)
	}
}
