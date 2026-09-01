package sentiment_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/sentiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

type fixedSignalStrategy struct {
	signal strategy.Signal
}

func (s fixedSignalStrategy) Name() string { return "Fixed" }

func (s fixedSignalStrategy) Analyze([]market.Candle) strategy.Signal { return s.signal }

func TestPostgresSentimentRuntimePath(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		_ = godotenv.Load("../../.env")
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("DATABASE_URL is not configured; set it or create backend/.env")
	}

	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	db := stdlib.OpenDB(*config)
	t.Cleanup(func() { _ = db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("connect to configured Postgres: %v", err)
	}

	repo := sentiment.NewPostgresRepository(db)
	lookup := sentiment.NewTimeLookup(repo, time.Hour)
	base := emptySentimentRange(t, ctx, db)
	runID := fmt.Sprintf("sentiment-runtime-check-%d", time.Now().UnixNano())
	insertedIDs := make([]string, 0, 8)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		for _, id := range insertedIDs {
			if _, err := db.ExecContext(cleanupCtx, `DELETE FROM sentiment_results WHERE news_id = $1`, id); err != nil {
				t.Errorf("clean up %s: %v", id, err)
			}
		}
	})

	seed := func(name, label string, confidence float64, publishedAt int64) {
		t.Helper()
		id := runID + "-" + name
		observation := sentiment.Observation{
			NewsID:       id,
			PublishedAt:  publishedAt,
			Sentiment:    label,
			Score:        confidence,
			ModelName:    "runtime-check",
			ModelVersion: "v1",
			AnalyzedAt:   time.Now().UnixMilli(),
		}
		if err := repo.Save(ctx, observation); err != nil {
			t.Fatalf("seed %s observation: %v", name, err)
		}
		insertedIDs = append(insertedIDs, id)
		// lookup caches ListSince results; without invalidating, a seed here
		// wouldn't be visible to the very next assertScore/strategy check
		// until the 1-minute TTL happened to expire.
		lookup.Invalidate()
	}

	assertScore := func(name string, candleTime int64, want float64) {
		t.Helper()
		got, err := lookup.FetchSentiment(ctx, candleTime)
		if err != nil {
			t.Fatalf("%s lookup: %v", name, err)
		}
		if math.Abs(got-want) > 1e-6 {
			t.Fatalf("%s score = %v, want %v", name, got, want)
		}
	}

	positiveCandle := base + (2 * time.Hour).Milliseconds()
	seed("positive", "POSITIVE", 0.9, positiveCandle-time.Minute.Milliseconds())
	assertScore("positive", positiveCandle, 0.9)

	negativeCandle := base + (2 * 24 * time.Hour).Milliseconds()
	seed("negative", "NEGATIVE", 0.9, negativeCandle-time.Minute.Milliseconds())
	assertScore("negative", negativeCandle, 0.1)

	neutralCandle := base + (4 * 24 * time.Hour).Milliseconds()
	seed("neutral", "NEUTRAL", 0.9, neutralCandle-time.Minute.Milliseconds())
	assertScore("neutral", neutralCandle, 0.5)

	futureCandle := base + (6 * 24 * time.Hour).Milliseconds()
	seed("future", "POSITIVE", 0.99, futureCandle+time.Minute.Milliseconds())
	if _, err := lookup.FetchSentiment(ctx, futureCandle); !errors.Is(err, sentiment.ErrNotFound) {
		t.Fatalf("future lookup error = %v, want ErrNotFound", err)
	}

	staleCandle := base + (8 * 24 * time.Hour).Milliseconds()
	seed("stale", "POSITIVE", 0.99, staleCandle-(2*time.Hour).Milliseconds())
	if _, err := lookup.FetchSentiment(ctx, staleCandle); !errors.Is(err, sentiment.ErrNotFound) {
		t.Fatalf("stale lookup error = %v, want ErrNotFound", err)
	}

	buyCandle := base + (10 * 24 * time.Hour).Milliseconds()
	seed("strategy-buy", "POSITIVE", 0.9, buyCandle-time.Minute.Milliseconds())
	buyStrategy := strategy.NewSentimentStrategy(fixedSignalStrategy{signal: strategy.Hold}, lookup, 0.8)
	if got := buyStrategy.Analyze([]market.Candle{{OpenTime: buyCandle}}); got != strategy.Buy {
		t.Fatalf("bullish strategy signal = %s, want BUY", got)
	}

	sellCandle := base + (12 * 24 * time.Hour).Milliseconds()
	seed("strategy-sell", "NEGATIVE", 0.9, sellCandle-time.Minute.Milliseconds())
	sellStrategy := strategy.NewSentimentStrategy(fixedSignalStrategy{signal: strategy.Hold}, lookup, 0.8)
	if got := sellStrategy.Analyze([]market.Candle{{OpenTime: sellCandle}}); got != strategy.Sell {
		t.Fatalf("bearish strategy signal = %s, want SELL", got)
	}

	fallbackCandle := base + (14 * 24 * time.Hour).Milliseconds()
	seed("strategy-stale", "POSITIVE", 0.99, fallbackCandle-(2*time.Hour).Milliseconds())
	fallbackStrategy := strategy.NewSentimentStrategy(fixedSignalStrategy{signal: strategy.Sell}, lookup, 0.8)
	if got := fallbackStrategy.Analyze([]market.Candle{{OpenTime: fallbackCandle}}); got != strategy.Sell {
		t.Fatalf("stale sentiment signal = %s, want base-strategy SELL", got)
	}

	missingCandle := base + (16 * 24 * time.Hour).Milliseconds()
	missingStrategy := strategy.NewSentimentStrategy(fixedSignalStrategy{signal: strategy.Buy}, lookup, 0.8)
	if got := missingStrategy.Analyze([]market.Candle{{OpenTime: missingCandle}}); got != strategy.Buy {
		t.Fatalf("missing sentiment signal = %s, want base-strategy BUY", got)
	}
}

func emptySentimentRange(t *testing.T, ctx context.Context, db *sql.DB) int64 {
	t.Helper()
	base := time.Date(2400, time.January, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	base += time.Now().UnixNano() % (100 * 365 * 24 * time.Hour).Milliseconds()
	for attempt := 0; attempt < 10; attempt++ {
		from := base - sentiment.DefaultMaxAge.Milliseconds()
		to := base + (16 * 24 * time.Hour).Milliseconds()
		var count int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sentiment_results WHERE published_at BETWEEN $1 AND $2`, from, to).Scan(&count); err != nil {
			t.Fatalf("inspect sentiment_results; apply backend/migrations/0002_sentiment_results.sql first: %v", err)
		}
		if count == 0 {
			return base
		}
		base += (30 * 24 * time.Hour).Milliseconds()
	}
	t.Fatal("could not find an isolated timestamp range for sentiment verification")
	return 0
}
