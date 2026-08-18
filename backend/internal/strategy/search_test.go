package strategy_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

func TestHandler_StartSearch(t *testing.T) {
	registry := strategy.NewRegistry()
	handler := strategy.NewHandler(registry)

	reqBody, _ := json.Marshal(strategy.StartSearchRequest{Count: 50})
	req, err := http.NewRequest(http.MethodPost, "/search/start", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.StartSearch(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response strategy.StartSearchResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}

	if response.Status != "STARTED" {
		t.Errorf("expected status STARTED, got %s", response.Status)
	}
	if response.Total != 50 {
		t.Errorf("expected total 50, got %d", response.Total)
	}
	if response.SearchID == "" {
		t.Errorf("expected non-empty search ID")
	}
}

func TestHandler_StartSearch_ValidationFail(t *testing.T) {
	registry := strategy.NewRegistry()
	handler := strategy.NewHandler(registry)

	// Test with invalid count
	reqBody, _ := json.Marshal(strategy.StartSearchRequest{Count: 0})
	req, err := http.NewRequest(http.MethodPost, "/search/start", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.StartSearch(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("expected bad request status, got %v", status)
	}
}
