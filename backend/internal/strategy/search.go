package strategy

import (
	"encoding/json"
	"errors"
	"net/http"
)

type StartSearchRequest struct {
	Count int `json:"count"`
}

func (r *StartSearchRequest) Validate() error {
	if r.Count <= 0 {
		return errors.New("count must be greater than 0")
	}
	if r.Count > 10000 {
		return errors.New("count cannot exceed 10000")
	}
	return nil
}

type StartSearchResponse struct {
	SearchID string `json:"searchId"`
	Status   string `json:"status"`
	Total    int    `json:"total"`
}

func (h *Handler) StartSearch(w http.ResponseWriter, r *http.Request) {
	var req StartSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// This is where we would generate Candidates and push them to the Job Queue
	// controlled by Person 3 (Experiment domain).
	// We'll generate a dummy search ID for now.
	searchID := generateID()

	// Example logic (Integration with Job Queue would go here):
	// generator := NewRandomGenerator(h.registry)
	// for i := 0; i < req.Count; i++ {
	// 	 candidate := generator.Generate()
	//   experimentQueue.Push(candidate)
	// }

	resp := StartSearchResponse{
		SearchID: searchID,
		Status:   "STARTED",
		Total:    req.Count,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
