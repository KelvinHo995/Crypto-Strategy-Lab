package strategy_test

import (
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

func TestMAStrategy_Analyze(t *testing.T) {
	s := strategy.NewMAStrategy(2, 4)

	if s.Name() != "MA" {
		t.Errorf("Expected name MA, got %s", s.Name())
	}

	tests := []struct {
		name    string
		candles []market.Candle
		want    strategy.Signal
	}{
		{
			name:    "not enough candles",
			candles: []market.Candle{{Close: 10}, {Close: 11}, {Close: 12}, {Close: 13}}, // needs 5 (LongWindow + 1)
			want:    strategy.Hold,
		},
		{
			name: "hold - no cross",
			candles: []market.Candle{
				{Close: 10}, {Close: 10}, {Close: 10}, {Close: 10}, // prev MAs: short=10, long=10
				{Close: 10}, // current MAs: short=10, long=10
			},
			want: strategy.Hold,
		},
		{
			name: "buy - short crosses above long",
			candles: []market.Candle{
				{Close: 10}, {Close: 10}, {Close: 10}, {Close: 10}, // prev MAs: short(10,10)=10, long(10x4)=10
				{Close: 14}, // current MAs: short(10,14)=12, long(10,10,10,14)=11 (12 > 11, cross above)
			},
			want: strategy.Buy,
		},
		{
			name: "sell - short crosses below long",
			candles: []market.Candle{
				{Close: 10}, {Close: 10}, {Close: 10}, {Close: 10}, // prev MAs: short=10, long=10
				{Close: 6}, // current MAs: short(10,6)=8, long(10,10,10,6)=9 (8 < 9, cross below)
			},
			want: strategy.Sell,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.Analyze(tt.candles); got != tt.want {
				t.Errorf("MAStrategy.Analyze() = %v, want %v", got, tt.want)
			}
		})
	}
}
