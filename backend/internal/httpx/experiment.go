package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
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

func startSearch(w http.ResponseWriter, r *http.Request) {
	var req StartSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func listExperiments(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func getExperiment(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
