package main

import (
	"log"
	"net/http"
	"os"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/httpx"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

func main() {
	registry := strategy.NewRegistry()
	registry.Register(strategy.NewMAStrategy(20, 50))
	registry.RegisterFactory("MA", strategy.MAFactory)
	registry.Register(strategy.NewRSIStrategy(14, 70.0, 30.0))
	registry.RegisterFactory("RSI", strategy.RSIFactory)
	registry.Register(strategy.NewBollingerStrategy(20, 2.0))
	registry.RegisterFactory("Bollinger", strategy.BollingerFactory)
	registry.Register(strategy.NewSRStrategy(20, 0.005))
	registry.RegisterFactory("SR", strategy.SRFactory)
	registry.Register(strategy.NewSMCStrategy(10))
	registry.RegisterFactory("SMC", strategy.SMCFactory)

	// Architecture Proof (ADR-0002): Assert expected strategies are registered
	expectedStrategies := 5 // MA, RSI, Bollinger, SR, SMC
	if len(registry.List()) != expectedStrategies {
		log.Fatalf("expected %d strategies, got %d", expectedStrategies, len(registry.List()))
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required — see backend/migrations/0001_init.sql and ADR-0012")
	}
	db, err := experiment.OpenDB(dsn)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer db.Close()
	repo := experiment.NewPostgresRepository(db)

	router := httpx.NewRouter(registry, repo)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
