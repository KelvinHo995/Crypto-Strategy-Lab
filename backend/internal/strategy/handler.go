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
	strategies := h.registry.List()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(strategies); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
