package strategy_test

import (
	"testing"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/strategy"
)

func TestStrategiesHoldForInvalidConfiguration(t *testing.T) {
	candles := make([]market.Candle, 30)
	for i := range candles {
		candles[i] = market.Candle{Open: 100, High: 101, Low: 99, Close: 100}
	}
	cases := []strategy.Strategy{
		strategy.NewMAStrategy(20, 10), strategy.NewRSIStrategy(0, 70, 30),
		strategy.NewBollingerStrategy(0, 2), strategy.NewSRStrategy(0, .01),
		strategy.NewSMCStrategy(0),
	}
	for _, s := range cases {
		if got := s.Analyze(candles); got != strategy.Hold {
			t.Errorf("%s returned %s, want HOLD", s.Name(), got)
		}
	}
}
