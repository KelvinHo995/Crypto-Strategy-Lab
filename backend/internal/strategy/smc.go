package strategy

import (
	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

var _ Strategy = (*SMCStrategy)(nil)

type SMCStrategy struct {
	SwingLookback int // Number of candles to look back to define a swing point
}

func NewSMCStrategy(swingLookback int) *SMCStrategy {
	return &SMCStrategy{
		SwingLookback: swingLookback,
	}
}

func SMCFactory(params map[string]any) Strategy {
	lookback, _ := toInt(params["smcLookback"], 10)
	return NewSMCStrategy(lookback)
}

func (s *SMCStrategy) Name() string {
	return "SMC"
}

func (s *SMCStrategy) Analyze(candles []market.Candle) Signal {
	if len(candles) < s.SwingLookback+2 {
		return Hold
	}

	endIndex := len(candles) - 1
	currentPrice := candles[endIndex].Close

	// Find recent swing high and low
	// A simple swing high is the highest high over the last SwingLookback candles (excluding current)
	// A simple swing low is the lowest low over the same period

	highestHigh := candles[endIndex-1].High
	lowestLow := candles[endIndex-1].Low

	startIndex := endIndex - s.SwingLookback
	if startIndex < 0 {
		startIndex = 0
	}

	for i := startIndex; i < endIndex; i++ {
		if candles[i].High > highestHigh {
			highestHigh = candles[i].High
		}
		if candles[i].Low < lowestLow {
			lowestLow = candles[i].Low
		}
	}

	// Break of Structure (BOS)
	if currentPrice > highestHigh {
		return Buy
	}

	if currentPrice < lowestLow {
		return Sell
	}

	return Hold
}
