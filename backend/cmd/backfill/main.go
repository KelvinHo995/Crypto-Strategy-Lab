package main

import (
	"context"
	"log"
	"os"
	"strings"
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
	symbol := os.Getenv("BACKFILL_SYMBOL")
	if symbol == "" {
		symbol = "BTCUSDT"
	}
	frames := os.Getenv("BACKFILL_TIMEFRAMES")
	if frames == "" {
		frames = "5m,15m,1h,4h"
	}
	to := time.Now().UnixMilli()
	from := time.Now().AddDate(-2, 0, 0).UnixMilli()
	provider := market.NewBinance(nil)
	repo := market.NewPostgresCandleRepository(db)
	for _, frame := range strings.Split(frames, ",") {
		frame = strings.TrimSpace(frame)
		candles, err := provider.FetchHistoricalCandles(ctx, symbol, frame, from, to)
		if err != nil {
			log.Fatalf("fetch %s: %v", frame, err)
		}
		if err := repo.Upsert(ctx, candles); err != nil {
			log.Fatalf("save %s: %v", frame, err)
		}
		log.Printf("backfilled %d %s candles", len(candles), frame)
	}
}
