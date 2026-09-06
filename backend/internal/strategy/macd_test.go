package strategy

import (
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

func TestMACDStrategyRejectsInvalidOrShortInput(t *testing.T) {
	if got := NewMACDStrategy(12, 26, 9).Analyze(make([]market.Candle, 34)); got != Hold {
		t.Fatalf("short input signal = %s, want HOLD", got)
	}
	if got := NewMACDStrategy(26, 12, 9).Analyze(make([]market.Candle, 50)); got != Hold {
		t.Fatalf("invalid config signal = %s, want HOLD", got)
	}
}

func TestMACDStrategyDetectsCrossovers(t *testing.T) {
	buyCandles := macdCandles(40, func(i int) float64 {
		if i == 39 {
			return 120
		}
		return 100
	})
	if got := NewMACDStrategy(12, 26, 9).Analyze(buyCandles); got != Buy {
		t.Fatalf("rising reversal signal = %s, want BUY", got)
	}

	sellCandles := macdCandles(40, func(i int) float64 {
		if i == 39 {
			return 80
		}
		return 100
	})
	if got := NewMACDStrategy(12, 26, 9).Analyze(sellCandles); got != Sell {
		t.Fatalf("falling reversal signal = %s, want SELL", got)
	}
}

func TestRegisterPluginRejectsDuplicate(t *testing.T) {
	registry := NewRegistry()
	if err := registry.RegisterPlugin(NewMACDPlugin(NewMACDStrategy(12, 26, 9))); err != nil {
		t.Fatal(err)
	}
	if err := registry.RegisterPlugin(NewMACDPlugin(NewMACDStrategy(8, 21, 5))); err == nil {
		t.Fatal("duplicate plugin registration succeeded")
	}
}

func macdCandles(count int, closeAt func(int) float64) []market.Candle {
	candles := make([]market.Candle, count)
	for i := range candles {
		candles[i] = market.Candle{Close: closeAt(i)}
	}
	return candles
}
