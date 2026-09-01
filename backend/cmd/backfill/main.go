package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")
	if os.Getenv("DATABASE_URL") == "" {
		_ = godotenv.Load("backend/.env")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	db, err := experiment.OpenDB(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	symbols, err := backfillSymbols(os.Getenv("BACKFILL_SYMBOLS"), os.Getenv("BACKFILL_SYMBOL"))
	if err != nil {
		log.Fatal(err)
	}
	frames := os.Getenv("BACKFILL_TIMEFRAMES")
	if frames == "" {
		frames = "5m,15m,1h,4h"
	}
	toTime := time.Now()
	fromTime := toTime.AddDate(-2, 0, 0)
	if rawDays := strings.TrimSpace(os.Getenv("BACKFILL_DAYS")); rawDays != "" {
		days, parseErr := strconv.Atoi(rawDays)
		if parseErr != nil || days < 1 || days > 3650 {
			log.Fatal("BACKFILL_DAYS must be between 1 and 3650")
		}
		fromTime = toTime.Add(-time.Duration(days) * 24 * time.Hour)
	}
	to := toTime.UnixMilli()
	from := fromTime.UnixMilli()
	pause := 150 * time.Millisecond
	if rawPause := strings.TrimSpace(os.Getenv("BACKFILL_REQUEST_PAUSE_MS")); rawPause != "" {
		milliseconds, parseErr := strconv.Atoi(rawPause)
		if parseErr != nil || milliseconds < 0 || milliseconds > 10_000 {
			log.Fatal("BACKFILL_REQUEST_PAUSE_MS must be between 0 and 10000")
		}
		pause = time.Duration(milliseconds) * time.Millisecond
	}
	client := &http.Client{Timeout: 30 * time.Second, Transport: &pacedTransport{base: http.DefaultTransport, pause: pause}}
	provider := market.NewBinance(client)
	repo := market.NewPostgresCandleRepository(db)
	for _, symbol := range symbols {
		for _, frame := range strings.Split(frames, ",") {
			frame = strings.TrimSpace(frame)
			candles, err := provider.FetchHistoricalCandles(ctx, symbol, frame, from, to)
			if err != nil {
				log.Fatalf("fetch %s/%s: %v", symbol, frame, err)
			}
			if err := repo.Upsert(ctx, candles); err != nil {
				log.Fatalf("save %s/%s: %v", symbol, frame, err)
			}
			log.Printf("backfilled %d %s/%s candles", len(candles), symbol, frame)
		}
	}
}

func backfillSymbols(rawMany, rawOne string) ([]string, error) {
	raw := rawMany
	if strings.TrimSpace(raw) == "" {
		raw = rawOne
	}
	if strings.TrimSpace(raw) == "" {
		raw = "BTCUSDT"
	}
	out := make([]string, 0)
	seen := make(map[string]struct{})
	invalid := make([]string, 0)
	for _, value := range strings.Split(raw, ",") {
		symbol := strings.ToUpper(strings.TrimSpace(value))
		if !market.IsSupportedSymbol(symbol) {
			invalid = append(invalid, symbol)
			continue
		}
		if _, exists := seen[symbol]; exists {
			continue
		}
		seen[symbol] = struct{}{}
		out = append(out, symbol)
	}
	if len(invalid) > 0 {
		return nil, fmt.Errorf("unsupported BACKFILL_SYMBOLS: %s", strings.Join(invalid, ", "))
	}
	return out, nil
}

type pacedTransport struct {
	base  http.RoundTripper
	pause time.Duration
	mu    sync.Mutex
	last  time.Time
}

func (t *pacedTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	t.mu.Lock()
	if wait := t.pause - time.Since(t.last); wait > 0 {
		timer := time.NewTimer(wait)
		select {
		case <-request.Context().Done():
			timer.Stop()
			t.mu.Unlock()
			return nil, request.Context().Err()
		case <-timer.C:
		}
	}
	t.last = time.Now()
	t.mu.Unlock()
	return t.base.RoundTrip(request)
}
