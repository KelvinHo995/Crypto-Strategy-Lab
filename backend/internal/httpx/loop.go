package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

type StartLoopRequest struct {
	Pair               string  `json:"pair"`
	TimeFrame          string  `json:"timeframe"`
	From               int64   `json:"from"`
	To                 int64   `json:"to"`
	Capital            float64 `json:"capital"`
	MaxCandidates      int     `json:"maxCandidates"`
	MaxDurationSeconds int     `json:"maxDurationSeconds"`
	NoImprovementLimit int     `json:"noImprovementLimit"`
}

func (r StartLoopRequest) Validate() error {
	if strings.TrimSpace(r.Pair) == "" {
		return errors.New("pair is required")
	}
	if !market.IsSupportedSymbol(r.Pair) {
		return errors.New("pair is not in the supported market catalog")
	}
	validTimeframes := map[string]bool{"5m": true, "15m": true, "1h": true, "4h": true}
	if !validTimeframes[r.TimeFrame] {
		return errors.New("timeframe must be one of 5m, 15m, 1h, 4h")
	}
	if r.From <= 0 || r.To <= 0 || r.To <= r.From {
		return errors.New("from/to must be a valid time range")
	}
	if r.Capital <= 0 {
		return errors.New("capital must be positive")
	}
	if r.MaxCandidates < 2 || r.MaxCandidates > 200 {
		return errors.New("maxCandidates must be between 2 and 200")
	}
	if r.MaxDurationSeconds != 0 && (r.MaxDurationSeconds < 60 || r.MaxDurationSeconds > 3600) {
		return errors.New("maxDurationSeconds must be between 60 and 3600 if set")
	}
	if r.NoImprovementLimit < 0 {
		return errors.New("noImprovementLimit cannot be negative")
	}
	return nil
}

type StartLoopResponse struct {
	SearchID      string `json:"searchId"`
	Status        string `json:"status"`
	MaxCandidates int    `json:"maxCandidates"`
}

// startLoop returns as soon as candles are loaded and the generator goroutine
// is launched — it does not wait for any candidate to finish, same
// fire-and-forget contract as startSearch.
func startLoop(gen strategy.StrategyGenerator, pool *experiment.WorkerPool, repo experiment.Repository, queue experiment.Queue, candleRepo market.CandleRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req StartLoopRequest
		if err := decodeJSON(w, r, &req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		req.Pair = strings.ToUpper(strings.TrimSpace(req.Pair))
		req.TimeFrame = strings.TrimSpace(req.TimeFrame)
		if err := req.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		candles := fixtureCandles(req.Pair, req.From, req.To)
		if candleRepo != nil {
			var err error
			candles, err = candleRepo.Range(r.Context(), req.Pair, req.TimeFrame, req.From, req.To)
			if err != nil {
				http.Error(w, "load historical candles", http.StatusInternalServerError)
				return
			}
			if len(candles) < 21 {
				http.Error(w, "insufficient historical candles; run backfill first", http.StatusUnprocessableEntity)
				return
			}
		}

		now := time.Now()
		searchID := experiment.NewJobID(now)
		params := experiment.LoopParams{
			SearchID: searchID, Pair: req.Pair, StartingCapital: req.Capital,
			Candles: candles, DatasetPeriod: fmt.Sprintf("%d-%d", req.From, req.To),
			MaxCandidates:      req.MaxCandidates,
			MaxDuration:        time.Duration(req.MaxDurationSeconds) * time.Second,
			NoImprovementLimit: req.NoImprovementLimit,
		}
		go experiment.RunSearchLoop(context.Background(), params, gen, pool, repo, queue)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(StartLoopResponse{SearchID: searchID, Status: "STARTED", MaxCandidates: req.MaxCandidates})
	}
}
