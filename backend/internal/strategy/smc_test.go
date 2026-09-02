package strategy

import (
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

func TestSMCStrategy_Analyze(t *testing.T) {
	strat := NewSMCStrategy(3)

	candles := []market.Candle{
		{High: 100, Low: 90},
		{High: 105, Low: 95}, // highestHigh = 105, lowestLow = 90
		{High: 102, Low: 92},
	}

	// Not enough data (requires Lookback + 2 = 5)
	if sig := strat.Analyze(candles); sig != Hold {
		t.Errorf("Expected Hold, got %v", sig)
	}

	candles = append(candles, market.Candle{High: 101, Low: 91}) // length 4, still Hold

	// Add 5th candle (BOS Bullish)
	candles = append(candles, market.Candle{Close: 106}) // > highestHigh (105)
	if sig := strat.Analyze(candles); sig != Buy {
		t.Errorf("Expected Buy for Bullish BOS, got %v", sig)
	}

	// Test Bearish BOS
	candles[4] = market.Candle{Close: 89} // < lowestLow (90)
	if sig := strat.Analyze(candles); sig != Sell {
		t.Errorf("Expected Sell for Bearish BOS, got %v", sig)
	}
}

func TestSMCStrategy_MinLookback(t *testing.T) {
	strat := NewSMCStrategy(10)
	if got := strat.MinLookback(); got != 12 {
		t.Fatalf("MinLookback() = %d, want 12 (SwingLookback+2)", got)
	}
}
