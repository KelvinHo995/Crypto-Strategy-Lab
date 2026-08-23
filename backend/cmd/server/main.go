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
	registry.Register(strategy.NewRSIStrategy(14, 70.0, 30.0))
	registry.Register(strategy.NewBollingerStrategy(20, 2.0))

	router := httpx.NewRouter(registry)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
