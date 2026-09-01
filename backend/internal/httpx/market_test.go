package httpx_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/httpx"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

func TestListCandlesNormalizesAndLimits(t *testing.T) {
	candles := &fakeCandleRepo{}
	router := httpx.NewRouterWithContext(context.Background(), newTestRegistry(), newFakeRepo(), httpx.Dependencies{Candles: candles})
	defer router.Close()
	srv := httptest.NewServer(router)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/candles?symbol=btcusdt&timeframe=5m&from=1&to=1000&limit=2")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	var got []market.Candle
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || candles.symbol != "BTCUSDT" {
		t.Fatalf("candles=%d symbol=%s", len(got), candles.symbol)
	}
}

func TestListCandlesRejectsUnsupportedTimeframe(t *testing.T) {
	router := httpx.NewRouterWithContext(context.Background(), newTestRegistry(), newFakeRepo(), httpx.Dependencies{Candles: &fakeCandleRepo{}})
	defer router.Close()
	req := httptest.NewRequest(http.MethodGet, "/candles?symbol=BTCUSDT&timeframe=1m&from=1&to=2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", w.Code)
	}
}

func TestListMarketsReturnsSupportedCatalog(t *testing.T) {
	router := httpx.NewRouter(newTestRegistry(), newFakeRepo())
	defer router.Close()
	request := httptest.NewRequest(http.MethodGet, "/markets", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d", response.Code)
	}
	var markets []market.Info
	if err := json.NewDecoder(response.Body).Decode(&markets); err != nil {
		t.Fatal(err)
	}
	if len(markets) != 8 || markets[0].Symbol != "BTCUSDT" || len(markets[0].Timeframes) != 4 {
		t.Fatalf("markets=%+v", markets)
	}
}

func TestListCandlesRejectsUnsupportedSymbol(t *testing.T) {
	router := httpx.NewRouterWithContext(context.Background(), newTestRegistry(), newFakeRepo(), httpx.Dependencies{Candles: &fakeCandleRepo{}})
	defer router.Close()
	request := httptest.NewRequest(http.MethodGet, "/candles?symbol=FAKEUSDT&timeframe=5m&from=1&to=2", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", response.Code)
	}
}
