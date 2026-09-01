package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

type StartSearchRequest struct {
	Pair       string   `json:"pair"`
	TimeFrame  string   `json:"timeframe"`
	From       int64    `json:"from"`
	To         int64    `json:"to"`
	Capital    float64  `json:"capital"`
	Strategies []string `json:"strategies"`
}

func (r StartSearchRequest) Validate() error {
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
	if len(r.Strategies) == 0 {
		return errors.New("at least one strategy is required")
	}
	seen := make(map[string]struct{}, len(r.Strategies))
	for _, name := range r.Strategies {
		if strings.TrimSpace(name) == "" {
			return errors.New("strategy names cannot be empty")
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("strategy %s is duplicated", name)
		}
		seen[name] = struct{}{}
	}
	return nil
}

type StartSearchResponse struct {
	SearchID string `json:"searchId"`
	Status   string `json:"status"`
}

// startSearch returns as soon as a validated, snapshotted candidate is queued.
// Production reads persisted market data; tests may omit the repository and use
// the deterministic fixture fallback.
func startSearch(registry *strategy.Registry, repo experiment.Repository, queue experiment.Queue, candleRepo market.CandleRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req StartSearchRequest
		if err := decodeJSON(w, r, &req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		req.Pair = strings.ToUpper(strings.TrimSpace(req.Pair))
		req.TimeFrame = strings.TrimSpace(req.TimeFrame)
		for i := range req.Strategies {
			req.Strategies[i] = strings.TrimSpace(req.Strategies[i])
		}
		if err := req.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		now := time.Now()
		id := experiment.NewJobID(now)
		candidate := strategy.CandidateStrategy{
			ID:         fmt.Sprintf("cand-%d", now.UnixNano()),
			Strategies: req.Strategies,
			Policy:     "majority",
		}
		_, err := strategy.BuildFromCandidate(registry, candidate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		versions := experiment.DefaultStrategyVersions(candidate.Strategies)
		candles := fixtureCandles(req.Pair, req.From, req.To)
		if candleRepo != nil {
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
		job := experiment.BacktestJob{
			ID: id, SearchID: id, SearchTotal: 1, Candidate: candidate,
			Candles: candles,
			Config: experiment.Config{Pair: req.Pair, StartingCapital: req.Capital,
				PositionSizePct: 1, StopLossPct: 0.02, TakeProfitPct: 0.04,
				FeePct: 0.001, SlippageBps: 5, Window: 20},
			DatasetPeriod:    fmt.Sprintf("%d-%d", req.From, req.To),
			StrategyVersions: versions, EnqueuedAt: now.UnixMilli(),
		}
		pending := experiment.Result{ID: id, SearchID: id, SearchTotal: 1, CandidateID: candidate.ID,
			Strategies: candidate.Strategies, Params: candidate.Params, Policy: candidate.Policy,
			StrategyVersions: versions, DatasetPeriod: job.DatasetPeriod,
			Status: "PENDING", CreatedAt: now.UnixMilli()}
		if err := repo.Save(r.Context(), pending); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := queue.Enqueue(r.Context(), job); err != nil {
			pending.Status = "FAILED"
			_ = repo.Save(r.Context(), pending)
			http.Error(w, "search queue unavailable", http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(StartSearchResponse{SearchID: id, Status: "STARTED"})
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}

func listExperiments(repo experiment.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results, err := repo.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if results == nil {
			results = []experiment.Result{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(experiment.Rank(results))
	}
}

func getExperiment(repo experiment.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		result, err := repo.Get(r.Context(), id)
		if errors.Is(err, experiment.ErrNotFound) {
			http.Error(w, "experiment not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
