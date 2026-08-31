package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/sentiment"
)

type SentimentService interface {
	AnalyzeAndStore(ctx context.Context, newsID, text string, publishedAt int64) (sentiment.Observation, error)
}

type AnalyzeSentimentRequest struct {
	NewsID      string `json:"newsId"`
	Text        string `json:"text"`
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

		observation, err := service.AnalyzeAndStore(r.Context(), req.NewsID, req.Text, req.PublishedAt)
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
