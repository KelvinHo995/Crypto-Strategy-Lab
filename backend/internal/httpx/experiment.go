package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/experiment"
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
	if r.Pair == "" {
		return errors.New("pair is required")
	}
	if r.TimeFrame == "" {
		return errors.New("time frame is required")
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
	return nil
}

type StartSearchResponse struct {
	SearchID string `json:"searchId"`
	Status   string `json:"status"`
}

// startSearch returns as soon as a validated, snapshotted candidate is queued.
// Candles remain a fixture until Market Data provides historical candles.
func startSearch(registry *strategy.Registry, repo experiment.Repository, queue experiment.Queue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req StartSearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
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
		job := experiment.BacktestJob{
			ID: id, Candidate: candidate,
			Candles: fixtureCandles(req.Pair, req.From, req.To),
			Config: experiment.Config{Pair: req.Pair, StartingCapital: req.Capital,
				PositionSizePct: 1, StopLossPct: 0.02, TakeProfitPct: 0.04,
				FeePct: 0.001, SlippageBps: 5, Window: 20},
			DatasetPeriod:    fmt.Sprintf("%d-%d", req.From, req.To),
			StrategyVersions: versions, EnqueuedAt: now.UnixMilli(),
		}
		pending := experiment.Result{ID: id, CandidateID: candidate.ID,
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

func listExperiments(repo experiment.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results, err := repo.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
