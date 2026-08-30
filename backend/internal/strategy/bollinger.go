package strategy

import (
	"math"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

var _ Strategy = (*BollingerStrategy)(nil)

type BollingerStrategy struct {
	Period           int
	StdDevMultiplier float64
}

func NewBollingerStrategy(period int, stdDevMultiplier float64) *BollingerStrategy {
	return &BollingerStrategy{
		Period:           period,
		StdDevMultiplier: stdDevMultiplier,
	}
}

func BollingerFactory(params map[string]any) Strategy {
	period, _ := toInt(params["bollingerPeriod"], 20)
	sd, _ := toFloat(params["bollingerStdDev"], 2.0)
	return NewBollingerStrategy(period, sd)
}

func (s *BollingerStrategy) Name() string {
	return "Bollinger"
}

func (s *BollingerStrategy) Analyze(candles []market.Candle) Signal {
	if s.Period <= 0 || s.StdDevMultiplier <= 0 || len(candles) < s.Period {
		return Hold
	}

	sma, stdDev := s.calculateStats(candles)

	upperBand := sma + (s.StdDevMultiplier * stdDev)
	lowerBand := sma - (s.StdDevMultiplier * stdDev)

	currentPrice := candles[len(candles)-1].Close

	if currentPrice < lowerBand {
		return Buy
	}
	if currentPrice > upperBand {
		return Sell
	}

	return Hold
}

func (s *BollingerStrategy) calculateStats(candles []market.Candle) (float64, float64) {
	endIndex := len(candles) - 1
	sum := 0.0

	// Calculate Mean (SMA)
	for i := endIndex - s.Period + 1; i <= endIndex; i++ {
		sum += candles[i].Close
	}
	mean := sum / float64(s.Period)

	// Calculate Variance
	varianceSum := 0.0
	for i := endIndex - s.Period + 1; i <= endIndex; i++ {
		diff := candles[i].Close - mean
		varianceSum += diff * diff
	}
	variance := varianceSum / float64(s.Period)
	stdDev := math.Sqrt(variance)

	return mean, stdDev
}
