package strategy

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

type Handler struct {
	registry *Registry
}

func NewHandler(registry *Registry) *Handler {
	return &Handler{
		registry: registry,
	}
}

func (h *Handler) ListStrategies(w http.ResponseWriter, r *http.Request) {
	strategies := h.registry.List()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(strategies); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// SignalRequest mirrors the composite shape used by /search/start (same
// instances/policy contract) — it's answering the same "what would this
// candidate do" question, just against the latest real candles instead of a
// historical range.
type SignalRequest struct {
	Pair      string             `json:"pair"`
	TimeFrame string             `json:"timeframe"`
	Instances []StrategyInstance `json:"instances"`
	Policy    string             `json:"policy"`
}

func (r SignalRequest) validate() error {
	if strings.TrimSpace(r.Pair) == "" {
		return errors.New("pair is required")
	}
	if !market.IsSupportedSymbol(r.Pair) {
		return errors.New("pair is not in the supported market catalog")
	}
	if !market.IsSupportedTimeframe(r.TimeFrame) {
		return errors.New("timeframe must be one of 5m, 15m, 1h, 4h")
	}
	if len(r.Instances) == 0 {
		return errors.New("at least one strategy instance is required")
	}
	for _, inst := range r.Instances {
		if strings.TrimSpace(inst.Type) == "" {
			return errors.New("strategy type cannot be empty")
		}
	}
	if r.Policy != "" && r.Policy != "majority" && r.Policy != "weighted" {
		return errors.New("policy must be majority or weighted")
	}
	return nil
}

type SignalResponse struct {
	Composite string   `json:"composite"`
	Signals   []string `json:"signals"` // aligned with the request's Instances order
}

// recentCandleLookbackWindow is a generous buffer safely covering the
// largest MinLookback of any registered strategy at any supported
// timeframe (200 candles at 4h = ~33 days).
const recentCandleLookbackWindow = 45 * 24 * time.Hour

// CurrentSignal computes each requested instance's real signal — and the
// composite's — against the most recently backfilled candles, using the
// exact same registry/combination-policy code path the real backtester
// runs, just against "now" instead of a historical range.
func (h *Handler) CurrentSignal(candles market.CandleRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SignalRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		req.Pair = strings.ToUpper(strings.TrimSpace(req.Pair))
		req.TimeFrame = strings.TrimSpace(req.TimeFrame)
		if err := req.validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		built := make([]Strategy, 0, len(req.Instances))
		weights := make([]float64, 0, len(req.Instances))
		lookback := 1
		for _, inst := range req.Instances {
			s, err := h.registry.BuildWithParams(inst.Type, inst.Params)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			built = append(built, s)
			weights = append(weights, inst.Weight)
			if la, ok := s.(LookbackAware); ok && la.MinLookback() > lookback {
				lookback = la.MinLookback()
			}
		}

		if candles == nil {
			http.Error(w, "market data unavailable", http.StatusServiceUnavailable)
			return
		}
		to := time.Now().UnixMilli()
		from := to - recentCandleLookbackWindow.Milliseconds()
		recent, err := candles.Range(r.Context(), req.Pair, req.TimeFrame, from, to)
		if err != nil {
			http.Error(w, "load recent candles", http.StatusInternalServerError)
			return
		}
		if len(recent) < lookback {
			http.Error(w, "insufficient recent candles; run backfill first", http.StatusUnprocessableEntity)
			return
		}
		window := recent[len(recent)-lookback:]

		signals := make([]Signal, len(built))
		signalStrings := make([]string, len(built))
		for i, s := range built {
			signals[i] = s.Analyze(window)
			signalStrings[i] = string(signals[i])
		}
		composite := resolveCombinationPolicy(req.Policy, weights).Combine(signals)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SignalResponse{Composite: string(composite), Signals: signalStrings})
	}
}
