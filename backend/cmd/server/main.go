package main

import (
	"log"
	"net/http"

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

	router := httpx.NewRouter(registry)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
