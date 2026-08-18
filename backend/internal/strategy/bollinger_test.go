package strategy_test

import (
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

func TestBollingerStrategy_Analyze(t *testing.T) {
	s := strategy.NewBollingerStrategy(20, 2.0)

	if s.Name() != "Bollinger" {
		t.Errorf("Expected name Bollinger, got %s", s.Name())
	}

	tests := []struct {
		name    string
		candles []market.Candle
		want    strategy.Signal
	}{
		{
			name:    "not enough candles",
			candles: make([]market.Candle, 10), // needs 20
			want:    strategy.Hold,
		},
		{
			name: "hold - price within bands",
			candles: func() []market.Candle {
				c := make([]market.Candle, 20)
				for i := range c {
					c[i] = market.Candle{Close: 100.0}
				}
				// stdDev = 0, upper/lower band = 100
				// current price = 100
				return c
			}(),
			want: strategy.Hold, // neither < nor >
		},
		{
			name: "buy - price below lower band",
			candles: func() []market.Candle {
				c := make([]market.Candle, 20)
				for i := range c {
					c[i] = market.Candle{Close: 100.0}
				}
				// change one of previous to create variance
				c[0] = market.Candle{Close: 120.0}
				
				// drop current price significantly
				c[19] = market.Candle{Close: 80.0}
				return c
			}(),
			want: strategy.Buy, 
		},
		{
			name: "sell - price above upper band",
			candles: func() []market.Candle {
				c := make([]market.Candle, 20)
				for i := range c {
					c[i] = market.Candle{Close: 100.0}
				}
				// change one of previous to create variance
				c[0] = market.Candle{Close: 80.0}
				
				// raise current price significantly
				c[19] = market.Candle{Close: 120.0}
				return c
			}(),
			want: strategy.Sell, 
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.Analyze(tt.candles); got != tt.want {
				t.Errorf("BollingerStrategy.Analyze() = %v, want %v", got, tt.want)
			}
		})
	}
}
