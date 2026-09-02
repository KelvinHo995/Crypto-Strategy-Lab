package strategy

import "github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"

var _ Strategy = (*RSIStrategy)(nil)

type RSIStrategy struct {
	Period              int
	OverboughtThreshold float64
	OversoldThreshold   float64
}

func NewRSIStrategy(period int, overbought, oversold float64) *RSIStrategy {
	return &RSIStrategy{
		Period:              period,
		OverboughtThreshold: overbought,
		OversoldThreshold:   oversold,
	}
}

func RSIFactory(params map[string]any) Strategy {
	period, _ := toInt(params["rsiPeriod"], 14)
	ob, _ := toFloat(params["rsiOverbought"], 70.0)
	os, _ := toFloat(params["rsiOversold"], 30.0)
	return NewRSIStrategy(period, ob, os)
}

func (s *RSIStrategy) Name() string {
	return "RSI"
}

func (s *RSIStrategy) MinLookback() int {
	return s.Period + 1
}

func (s *RSIStrategy) Analyze(candles []market.Candle) Signal {
	if s.Period <= 0 || s.OversoldThreshold < 0 || s.OverboughtThreshold > 100 || s.OversoldThreshold >= s.OverboughtThreshold || len(candles) <= s.Period {
		return Hold
	}

	rsi := s.calculateRSI(candles)

	if rsi < s.OversoldThreshold {
		return Buy
	}
	if rsi > s.OverboughtThreshold {
		return Sell
	}

	return Hold
}

func (s *RSIStrategy) calculateRSI(candles []market.Candle) float64 {
	// Need at least Period + 1 candles to have 'Period' number of changes
	if len(candles) <= s.Period {
		return 50.0 // Default neutral
	}

	// Calculate initial SMA of gains and losses for the first 'Period'
	sumGain := 0.0
	sumLoss := 0.0

	// We only use the last (Period * 2) candles to save calculation time if the array is huge,
	// but to be precise, RSI typically uses all available historical data to calculate smoothed moving average.
	// For simplicity in this lab, we'll calculate it from the beginning of the available slice.

	// Start index to ensure we have enough data to calculate smoothed RSI
	// We'll just calculate from start of array.
	for i := 1; i <= s.Period; i++ {
		change := candles[i].Close - candles[i-1].Close
		if change > 0 {
			sumGain += change
		} else {
			sumLoss -= change
		}
	}

	avgGain := sumGain / float64(s.Period)
	avgLoss := sumLoss / float64(s.Period)

	// Wilder's Smoothing for the rest of the candles
	for i := s.Period + 1; i < len(candles); i++ {
		change := candles[i].Close - candles[i-1].Close
		gain := 0.0
		loss := 0.0
		if change > 0 {
			gain = change
		} else {
			loss = -change
		}

		avgGain = (avgGain*float64(s.Period-1) + gain) / float64(s.Period)
		avgLoss = (avgLoss*float64(s.Period-1) + loss) / float64(s.Period)
	}

	if avgLoss == 0 {
		return 100.0
	}

	rs := avgGain / avgLoss
	rsi := 100.0 - (100.0 / (1.0 + rs))

	return rsi
}
