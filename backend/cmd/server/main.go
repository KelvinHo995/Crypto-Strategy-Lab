package main

import (
	"log"
	"net/http"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

func main() {
	registry := strategy.NewRegistry()
	registry.Register(strategy.NewMAStrategy(20, 50))
	registry.Register(strategy.NewRSIStrategy(14, 70.0, 30.0))
	registry.Register(strategy.NewBollingerStrategy(20, 2.0))

	strategyHandler := strategy.NewHandler(registry)

	// Since go 1.22 we can use method-based routing
	http.HandleFunc("GET /strategies", strategyHandler.ListStrategies)
	http.HandleFunc("POST /search/start", strategyHandler.StartSearch)
	
	http.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
