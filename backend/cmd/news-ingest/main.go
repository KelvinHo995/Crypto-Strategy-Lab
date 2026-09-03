package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/news"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/sentiment"
	"github.com/joho/godotenv"
)

const defaultNewsLookback = 24 * time.Hour

func main() {
	loadLocalEnv()

	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	feedURLs := splitFeedURLs(os.Getenv("NEWS_RSS_FEEDS"))
	if len(feedURLs) == 0 {
		log.Fatal("NEWS_RSS_FEEDS is required (comma-separated RSS URLs)")
	}

	lookback := defaultNewsLookback
	if raw := strings.TrimSpace(os.Getenv("NEWS_LOOKBACK")); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil || parsed <= 0 {
			log.Fatal("NEWS_LOOKBACK must be a positive Go duration such as 24h")
		}
		lookback = parsed
	}
	sentimentURL := strings.TrimSpace(os.Getenv("SENTIMENT_SERVICE_URL"))
	if sentimentURL == "" {
		sentimentURL = "http://localhost:8000"
	}

	db, err := experiment.OpenDB(dsn)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer db.Close()

	feedClient := &http.Client{Timeout: 10 * time.Second}
	analyzerClient := sentiment.NewClient(sentimentURL, &http.Client{Timeout: 5 * time.Second})
	service := sentiment.NewService(analyzerClient, sentiment.NewPostgresRepository(db))
	provider := news.NewRSSNewsProvider(feedURLs, feedClient)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	since := time.Now().Add(-lookback).UnixMilli()
	observations, ingestErr := service.IngestFromProvider(ctx, provider, since)
	successfulFeeds := len(feedURLs)
	var fetchErr *news.RSSFetchError
	if errors.As(ingestErr, &fetchErr) {
		successfulFeeds -= fetchErr.FailedFeedCount()
	}
	log.Printf(
		"RSS ingestion: configured_feeds=%d successful_feeds=%d analyzed_articles=%d",
		len(feedURLs), successfulFeeds, len(observations),
	)
	if ingestErr != nil {
		if fetchErr != nil && successfulFeeds > 0 {
			log.Printf("ingestion completed with a source warning: %v", ingestErr)
		} else {
			log.Fatal(ingestErr)
		}
	}
}

func loadLocalEnv() {
	for _, path := range []string{".env", "backend/.env"} {
		err := godotenv.Load(path)
		if err == nil {
			return
		}
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("load %s: %v", path, err)
			return
		}
	}
}

func splitFeedURLs(raw string) []string {
	feedURLs := make([]string, 0)
	seen := make(map[string]struct{})
	for _, value := range strings.Split(raw, ",") {
		feedURL := strings.TrimSpace(value)
		if feedURL == "" {
			continue
		}
		if _, duplicate := seen[feedURL]; duplicate {
			continue
		}
		seen[feedURL] = struct{}{}
		feedURLs = append(feedURLs, feedURL)
	}
	return feedURLs
}
