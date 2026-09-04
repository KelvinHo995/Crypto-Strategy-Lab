package strategy_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

type fakeCandleRepo struct {
	candles []market.Candle
	err     error
}

func (f *fakeCandleRepo) Upsert(ctx context.Context, candles []market.Candle) error { return nil }

func (f *fakeCandleRepo) Range(ctx context.Context, symbol, timeframe string, from, to int64) ([]market.Candle, error) {
	return f.candles, f.err
}

func repoWithCandles(n int) *fakeCandleRepo {
	candles := make([]market.Candle, n)
	for i := range candles {
		candles[i] = market.Candle{Symbol: "BTCUSDT", Timeframe: "1h", OpenTime: int64(i), Open: 100, High: 101, Low: 99, Close: 100, Volume: 1}
	}
	return &fakeCandleRepo{candles: candles}
}

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

func TestHandler_CurrentSignal_ComputesRealCompositeFromRegisteredStrategies(t *testing.T) {
	registry := strategy.NewRegistry()
	registry.Register(&DummyStrategy{AlwaysSignal: strategy.Buy})
	registry.RegisterFactory("DummySell", func(map[string]any) strategy.Strategy {
		return &DummyStrategy{AlwaysSignal: strategy.Sell}
	})
	handler := strategy.NewHandler(registry)

	body := `{"pair":"BTCUSDT","timeframe":"1h","policy":"weighted","instances":[` +
		`{"type":"Dummy","weight":0.8},{"type":"DummySell","weight":0.2}]}`
	req := httptest.NewRequest(http.MethodPost, "/strategies/signal", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	handler.CurrentSignal(repoWithCandles(10))(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%s", rr.Code, rr.Body.String())
	}
	var resp strategy.SignalResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Signals) != 2 || resp.Signals[0] != "BUY" || resp.Signals[1] != "SELL" {
		t.Fatalf("Signals = %v, want [BUY SELL] (aligned with request order)", resp.Signals)
	}
	// weighted: 0.8*Buy - 0.2*Sell > 0 -> Buy wins
	if resp.Composite != "BUY" {
		t.Fatalf("Composite = %q, want BUY (0.8-weighted Buy should outweigh 0.2-weighted Sell)", resp.Composite)
	}
}

func TestHandler_CurrentSignal_RejectsUnsupportedMarketOrTimeframe(t *testing.T) {
	registry := strategy.NewRegistry()
	registry.Register(&DummyStrategy{AlwaysSignal: strategy.Hold})
	handler := strategy.NewHandler(registry)

	for _, body := range []string{
		`{"pair":"NOTREAL","timeframe":"1h","instances":[{"type":"Dummy"}]}`,
		`{"pair":"BTCUSDT","timeframe":"3m","instances":[{"type":"Dummy"}]}`,
		`{"pair":"BTCUSDT","timeframe":"1h","instances":[]}`,
		`{"pair":"BTCUSDT","timeframe":"1h","instances":[{"type":"NOPE"}]}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/strategies/signal", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()
		handler.CurrentSignal(repoWithCandles(10))(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400 for %s", rr.Code, body)
		}
	}
}

func TestHandler_CurrentSignal_InsufficientRecentCandles(t *testing.T) {
	registry := strategy.NewRegistry()
	registry.Register(&DummyStrategy{AlwaysSignal: strategy.Hold})
	handler := strategy.NewHandler(registry)

	body := `{"pair":"BTCUSDT","timeframe":"1h","instances":[{"type":"Dummy"}]}`
	req := httptest.NewRequest(http.MethodPost, "/strategies/signal", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()

	handler.CurrentSignal(&fakeCandleRepo{candles: nil})(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422 when not enough recent candles are backfilled", rr.Code)
	}
}
