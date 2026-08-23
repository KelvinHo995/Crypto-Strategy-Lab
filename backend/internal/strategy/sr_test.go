package strategy

import (
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

func TestSRStrategy_Analyze(t *testing.T) {
	strat := NewSRStrategy(3, 0.01) // 1% tolerance

	candles := []market.Candle{
		{Close: 100, Low: 90, High: 110},
		{Close: 105, Low: 95, High: 115},
		{Close: 110, Low: 100, High: 120},
	}

	// Not enough candles
	if sig := strat.Analyze(candles); sig != Hold {
		t.Errorf("Expected Hold due to insufficient data, got %v", sig)
	}

	// Add 4th candle near support (minLow is 90)
	candles = append(candles, market.Candle{Close: 90.5}) // (90.5-90)/90 = 0.5% <= 1%
	if sig := strat.Analyze(candles); sig != Buy {
		t.Errorf("Expected Buy near support, got %v", sig)
	}

	// Add 5th candle near resistance (maxHigh is 120 from last 3 candles: 95,100,90.5 -> wait, last 3 candles (indices 1,2,3) maxHigh is 120 (at index 2).
	// Let's test resistance
	candles[3] = market.Candle{Close: 119.5, High: 119.5, Low: 110} // replace 4th candle to not trigger Buy
	candles = append(candles, market.Candle{Close: 119.0}) // near maxHigh 120 (from index 2)
	// indices 1, 2, 3 highs: 115, 120, 119.5. maxHigh = 120
	// current price 119.0. (120-119)/120 = 0.83% <= 1%
	if sig := strat.Analyze(candles); sig != Sell {
		t.Errorf("Expected Sell near resistance, got %v", sig)
	}
}
