package main

import (
	"log"
	"net/http"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", handleWS)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error:", err)
		return
	}
	defer clientConn.Close()

	candles := make(chan market.Candle)
	binance := &market.Binance{}

	go func() {
		if err := binance.StreamLiveCandles("btcusdt", "5m", candles); err != nil {
			log.Println("binance stream error:", err)
			close(candles)
		}
	}()

	for c := range candles {
		if err := clientConn.WriteJSON(c); err != nil {
			log.Println("client write error:", err)
			return
		}
	}
}
