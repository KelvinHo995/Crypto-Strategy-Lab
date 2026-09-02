package strategy

import (
	"math"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

var _ Strategy = (*SRStrategy)(nil)

type SRStrategy struct {
	Window    int
	Tolerance float64 // e.g. 0.005 for 0.5%
}

func NewSRStrategy(window int, tolerance float64) *SRStrategy {
	return &SRStrategy{
		Window:    window,
		Tolerance: tolerance,
	}
}

func SRFactory(params map[string]any) Strategy {
	window, _ := toInt(params["srWindow"], 20)
	tolerance, _ := toFloat(params["srTolerance"], 0.005)
	return NewSRStrategy(window, tolerance)
}

func (s *SRStrategy) Name() string {
	return "SR"
}

func (s *SRStrategy) MinLookback() int {
	return s.Window + 1
}

func (s *SRStrategy) Analyze(candles []market.Candle) Signal {
	if s.Window <= 0 || s.Tolerance < 0 || len(candles) <= s.Window {
		return Hold
	}

	endIndex := len(candles) - 1
	currentPrice := candles[endIndex].Close

	minLow := math.MaxFloat64
	maxHigh := -math.MaxFloat64

	startIndex := endIndex - s.Window
	if startIndex < 0 {
		startIndex = 0
	}

	for i := startIndex; i < endIndex; i++ {
		if candles[i].Low < minLow {
			minLow = candles[i].Low
		}
		if candles[i].High > maxHigh {
			maxHigh = candles[i].High
		}
	}

	if minLow > 0 && math.Abs(currentPrice-minLow)/minLow <= s.Tolerance {
		return Buy
	}

	if maxHigh > 0 && math.Abs(currentPrice-maxHigh)/maxHigh <= s.Tolerance {
		return Sell
	}

	return Hold
}
