package main

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/news"
)

func TestDemoRSSFeedExercisesRealParser(t *testing.T) {
	now := time.Date(2026, time.September, 6, 12, 0, 0, 0, time.UTC)
	server := startDemoRSSFeed(now)
	defer server.Close()

	provider := news.NewRSSNewsProvider([]string{server.URL}, &http.Client{Timeout: time.Second})
	items, err := provider.Fetch(context.Background(), now.Add(-time.Hour).UnixMilli())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("items=%d, want positive/negative/neutral fixtures", len(items))
	}
	if items[0].Source != "Crypto Strategy Lab Demo RSS" || !strings.Contains(items[0].Title, "bullish") {
		t.Fatalf("unexpected positive fixture: %+v", items[0])
	}
	if !strings.Contains(items[1].Text, "fraud risk") {
		t.Fatalf("unexpected negative fixture: %+v", items[1])
	}
}
