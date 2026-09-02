package strategy_test

import (
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

func TestRSIStrategy_Analyze(t *testing.T) {
	s := strategy.NewRSIStrategy(14, 70.0, 30.0)

	if s.Name() != "RSI" {
		t.Errorf("Expected name RSI, got %s", s.Name())
	}

	tests := []struct {
		name    string
		candles []market.Candle
		want    strategy.Signal
	}{
		{
			name:    "not enough candles",
			candles: make([]market.Candle, 10), // needs > 14
			want:    strategy.Hold,
		},
		{
			name: "hold - neutral RSI",
			candles: func() []market.Candle {
				c := make([]market.Candle, 20)
				// Alternate slightly up and down to keep RSI around 50
				val := 100.0
				for i := range c {
					if i%2 == 0 {
						val += 1.0
					} else {
						val -= 1.0
					}
					c[i] = market.Candle{Close: val}
				}
				return c
			}(),
			want: strategy.Hold,
		},
		{
			name: "buy - oversold RSI",
			candles: func() []market.Candle {
				c := make([]market.Candle, 20)
				// Consistently dropping to make RSI low
				val := 100.0
				for i := range c {
					val -= 2.0
					c[i] = market.Candle{Close: val}
				}
				return c
			}(),
			want: strategy.Buy, // RSI should be 0 (< 30)
		},
		{
			name: "sell - overbought RSI",
			candles: func() []market.Candle {
				c := make([]market.Candle, 20)
				// Consistently rising to make RSI high
				val := 100.0
				for i := range c {
					val += 2.0
					c[i] = market.Candle{Close: val}
				}
				return c
			}(),
			want: strategy.Sell, // RSI should be 100 (> 70)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.Analyze(tt.candles); got != tt.want {
				t.Errorf("RSIStrategy.Analyze() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRSIStrategy_MinLookback(t *testing.T) {
	s := strategy.NewRSIStrategy(14, 70.0, 30.0)
	if got := s.MinLookback(); got != 15 {
		t.Fatalf("MinLookback() = %d, want 15 (Period+1)", got)
	}
}
