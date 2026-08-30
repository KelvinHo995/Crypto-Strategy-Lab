package market_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

func TestBinanceFetchHistoricalCandles(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/klines" {
			t.Errorf("path=%s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[[1000,"100","110","90","105","12.5",1500],[2000,"105","115","95","110","13",2500]]`))
	}))
	defer s.Close()
	b := market.NewBinance(s.Client())
	b.BaseURL = s.URL
	c, err := b.FetchHistoricalCandles(context.Background(), "BTCUSDT", "5m", 1, 3000)
	if err != nil {
		t.Fatal(err)
	}
	if len(c) != 2 || c[0].Close != 105 || !c[0].IsClosed {
		t.Fatalf("candles=%+v", c)
	}
}
func TestBinanceRejectsInvalidRequest(t *testing.T) {
	b := market.NewBinance(nil)
	if _, err := b.FetchHistoricalCandles(context.Background(), "", "1m", 0, 0); err == nil {
		t.Fatal("invalid request accepted")
	}
}
