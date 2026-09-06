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
	Pair      string                      `json:"pair"`
	TimeFrame string                      `json:"timeframe"`
	From      int64                       `json:"from"`
	To        int64                       `json:"to"`
	Capital   float64                     `json:"capital"`
	Instances []strategy.StrategyInstance `json:"instances"`
	Policy    string                      `json:"policy"`
	Fee       float64                     `json:"fee"`      // percent, e.g. 0.1 = 0.1%; 0 uses the project default
	Slippage  float64                     `json:"slippage"` // bps, e.g. 5 = 5bps; 0 uses the project default
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
	if len(r.Instances) == 0 {
		return errors.New("at least one strategy instance is required")
	}
	// Duplicate types are allowed on purpose — e.g. MA(20) and MA(50) as two
	// independently configured instances in the same composite.
	for _, inst := range r.Instances {
		if strings.TrimSpace(inst.Type) == "" {
			return errors.New("strategy type cannot be empty")
		}
	}
	if r.Policy != "" && r.Policy != "majority" && r.Policy != "weighted" {
		return errors.New("policy must be majority or weighted")
	}
	return validateFeeSlippage(r.Fee, r.Slippage)
}

const (
	defaultFeePct      = 0.001 // 0.1%
	defaultSlippageBps = 5
)

// resolveFeeSlippage converts a request's percent-fee/bps-slippage into the
// backtester's internal units, defaulting to the project's baseline cost
// assumptions when the caller omits them (0 means "not set", not "free").
func resolveFeeSlippage(feePercent, slippageBps float64) (feePct, resolvedSlippageBps float64) {
	feePct = defaultFeePct
	if feePercent != 0 {
		feePct = feePercent / 100
	}
	resolvedSlippageBps = defaultSlippageBps
	if slippageBps != 0 {
		resolvedSlippageBps = slippageBps
	}
	return feePct, resolvedSlippageBps
}

func validateFeeSlippage(feePercent, slippageBps float64) error {
	if feePercent < 0 || feePercent > 2 {
		return errors.New("fee must be between 0 and 2 percent")
	}
	if slippageBps < 0 || slippageBps > 200 {
		return errors.New("slippage must be between 0 and 200 bps")
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
		for i := range req.Instances {
			req.Instances[i].Type = strings.TrimSpace(req.Instances[i].Type)
		}
		if err := req.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		policy := req.Policy
		if policy == "" {
			policy = "majority"
		}
		now := time.Now()
		id := experiment.NewJobID(now)
		candidate := strategy.CandidateStrategy{
			ID:        fmt.Sprintf("cand-%d", now.UnixNano()),
			Instances: req.Instances,
			Policy:    policy,
		}
		_, err := strategy.BuildFromCandidate(registry, candidate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		versions, err := registry.VersionsFor(candidate.Instances)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		candles := fixtureCandles(req.Pair, req.From, req.To)
		if candleRepo != nil {
			candles, err = candleRepo.Range(r.Context(), req.Pair, req.TimeFrame, req.From, req.To)
			if err != nil {
				http.Error(w, "load historical candles", http.StatusInternalServerError)
				return
			}
			if len(candles) < experiment.MinCandlesForBacktest {
				http.Error(w, "insufficient historical candles; run backfill first", http.StatusUnprocessableEntity)
				return
			}
		}
		feePct, slippageBps := resolveFeeSlippage(req.Fee, req.Slippage)
		job := experiment.BacktestJob{
			ID: id, SearchID: id, SearchTotal: 1, Candidate: candidate,
			Pair: req.Pair, Timeframe: req.TimeFrame, From: req.From, To: req.To,
			Candles: candles,
			Config: experiment.Config{Pair: req.Pair, StartingCapital: req.Capital,
				PositionSizePct: 1, StopLossPct: 0.02, TakeProfitPct: 0.04,
				FeePct: feePct, SlippageBps: slippageBps, AllowShort: true}, // Window is sized per-candidate by the worker (see worker.go)
			DatasetPeriod:    fmt.Sprintf("%d-%d", req.From, req.To),
			StrategyVersions: versions, EnqueuedAt: now.UnixMilli(),
		}
		pending := experiment.Result{ID: id, SearchID: id, SearchTotal: 1, CandidateID: candidate.ID,
			Pair: req.Pair, Timeframe: req.TimeFrame,
			Instances: candidate.Instances, Policy: candidate.Policy,
			StrategyVersions: versions, DatasetPeriod: job.DatasetPeriod,
			Status: "PENDING", CreatedAt: now.UnixMilli()}
		if durable, ok := queue.(experiment.PendingQueue); ok {
			if err := durable.EnqueuePending(r.Context(), job, pending); err != nil {
				http.Error(w, "search queue unavailable", http.StatusServiceUnavailable)
				return
			}
			if invalidator, ok := repo.(interface{ Invalidate() }); ok {
				invalidator.Invalidate()
			}
		} else {
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

// getExperimentTrades returns the real trade-by-trade history for one
// completed backtest. Empty (not an error) for experiments that predate
// trade persistence, or that never completed.
func getExperimentTrades(repo experiment.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if _, err := repo.Get(r.Context(), id); errors.Is(err, experiment.ErrNotFound) {
			http.Error(w, "experiment not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		trades, err := repo.ListTrades(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if trades == nil {
			trades = []experiment.Trade{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(trades)
	}
}
