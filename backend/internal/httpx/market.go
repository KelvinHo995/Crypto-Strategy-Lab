package httpx

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

func listCandles(repo market.CandleRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if repo == nil {
			http.Error(w, "market data unavailable", http.StatusServiceUnavailable)
			return
		}
		symbol := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("symbol")))
		timeframe := strings.TrimSpace(r.URL.Query().Get("timeframe"))
		from, fromErr := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
		to, toErr := strconv.ParseInt(r.URL.Query().Get("to"), 10, 64)
		if symbol == "" || fromErr != nil || toErr != nil || from <= 0 || to <= from {
			http.Error(w, "symbol and valid from/to are required", http.StatusBadRequest)
			return
		}
		if timeframe != "5m" && timeframe != "15m" && timeframe != "1h" && timeframe != "4h" {
			http.Error(w, "unsupported timeframe", http.StatusBadRequest)
			return
		}
		limit := 1000
		if raw := r.URL.Query().Get("limit"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err != nil || parsed < 1 || parsed > 5000 {
				http.Error(w, "limit must be between 1 and 5000", http.StatusBadRequest)
				return
			} else {
				limit = parsed
			}
		}
		intervals := map[string]int64{"5m": 300_000, "15m": 900_000, "1h": 3_600_000, "4h": 14_400_000}
		minimumFrom := to - int64(limit)*intervals[timeframe]
		if from < minimumFrom {
			from = minimumFrom
		}
		candles, err := repo.Range(r.Context(), symbol, timeframe, from, to)
		if err != nil {
			http.Error(w, "load market data", http.StatusInternalServerError)
			return
		}
		if len(candles) > limit {
			candles = candles[len(candles)-limit:]
		}
		if candles == nil {
			candles = []market.Candle{}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "private, max-age="+strconv.Itoa(int((30*time.Second).Seconds())))
		_ = json.NewEncoder(w).Encode(candles)
	}
}
