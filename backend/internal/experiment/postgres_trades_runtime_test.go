package experiment_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/joho/godotenv"
)

func TestPostgresRepository_SaveAndListTrades(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		_ = godotenv.Load("../../.env")
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("DATABASE_URL is not configured")
	}
	db, err := experiment.OpenDB(dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := experiment.NewPostgresRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	id := "trade-history-runtime-test"
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		_, _ = db.ExecContext(cleanup, `DELETE FROM experiments WHERE id=$1`, id)
	})

	result := experiment.Result{
		ID: id, SearchID: id, SearchTotal: 1, CandidateID: "cand-1",
		Policy: "majority", Status: "COMPLETED", CreatedAt: time.Now().UnixMilli(),
	}
	if err := repo.Save(ctx, result); err != nil {
		t.Fatalf("seed experiment: %v", err)
	}

	trades := []experiment.Trade{
		{
			Pair: "BTCUSDT", EntryTime: 1000, Direction: experiment.Long,
			VolumeUSD: 1000, EntryPrice: 100, StopLoss: 95, TakeProfit: 110,
			ExitPrice: 108, ExitTime: 2000, TransactionCost: 1, Slippage: 0.5, Profit: 79.5,
		},
		{
			Pair: "BTCUSDT", EntryTime: 3000, Direction: experiment.Short,
			VolumeUSD: 1000, EntryPrice: 110, StopLoss: 115, TakeProfit: 100,
			ExitPrice: 102, ExitTime: 4000, TransactionCost: 1, Slippage: 0.5, Profit: 70.5,
		},
	}
	if err := repo.SaveTrades(ctx, id, trades); err != nil {
		t.Fatalf("SaveTrades: %v", err)
	}

	got, err := repo.ListTrades(ctx, id)
	if err != nil {
		t.Fatalf("ListTrades: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d trades, want 2", len(got))
	}
	if got[0].EntryTime != 1000 || got[0].Direction != experiment.Long || got[0].Profit != 79.5 {
		t.Errorf("trade[0] = %+v, want the seeded LONG trade", got[0])
	}
	if got[1].EntryTime != 3000 || got[1].Direction != experiment.Short || got[1].Profit != 70.5 {
		t.Errorf("trade[1] = %+v, want the seeded SHORT trade", got[1])
	}

	// A second SaveTrades call must fully replace, not append — otherwise a
	// retried worker save would silently double every trade.
	if err := repo.SaveTrades(ctx, id, trades[:1]); err != nil {
		t.Fatalf("SaveTrades (replace): %v", err)
	}
	got, err = repo.ListTrades(ctx, id)
	if err != nil {
		t.Fatalf("ListTrades after replace: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d trades after replace, want 1 (SaveTrades must replace, not append)", len(got))
	}
}
