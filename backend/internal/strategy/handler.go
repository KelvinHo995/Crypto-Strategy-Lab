package strategy

import (
	"encoding/json"
	"net/http"
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
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	strategies := h.registry.List()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(strategies); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
