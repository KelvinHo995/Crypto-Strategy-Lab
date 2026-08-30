package sentiment_test

import (
	"context"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/sentiment"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientAnalyze(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"newsId":"n1","sentiment":"POSITIVE","score":0.9,"model":{"name":"crypto-lexicon","version":"v1"},"createdAt":1}`))
	}))
	defer s.Close()
	r, err := sentiment.NewClient(s.URL, s.Client()).Analyze(context.Background(), "n1", "bullish rally")
	if err != nil {
		t.Fatal(err)
	}
	if r.Model.Version != "v1" || r.Score != .9 {
		t.Fatalf("result=%+v", r)
	}
}
func TestClientServiceDown(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "down", 503) }))
	defer s.Close()
	if _, err := sentiment.NewClient(s.URL, s.Client()).Analyze(context.Background(), "n1", "text"); err == nil {
		t.Fatal("service error ignored")
	}
}
