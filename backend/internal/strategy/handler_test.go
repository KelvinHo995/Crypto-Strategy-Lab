package strategy_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

func TestHandler_ListStrategies(t *testing.T) {
	registry := strategy.NewRegistry()
	registry.Register(&DummyStrategy{AlwaysSignal: strategy.Hold})

	handler := strategy.NewHandler(registry)

	req, err := http.NewRequest(http.MethodGet, "/strategies", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ListStrategies(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response []string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if len(response) != 1 || response[0] != "Dummy" {
		t.Errorf("handler returned unexpected body: got %v", response)
	}
}

func TestHandler_ListStrategies_MethodNotAllowed(t *testing.T) {
	registry := strategy.NewRegistry()
	handler := strategy.NewHandler(registry)

	req, err := http.NewRequest(http.MethodPost, "/strategies", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ListStrategies(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code for POST: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}
