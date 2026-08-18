package strategy

import "github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"

var _ Strategy = (*MAStrategy)(nil)

type MAStrategy struct {
	ShortWindow int
	LongWindow  int
}

func NewMAStrategy(shortWindow, longWindow int) *MAStrategy {
	return &MAStrategy{
		ShortWindow: shortWindow,
		LongWindow:  longWindow,
	}
}

func (s *MAStrategy) Name() string {
	return "MA"
}

func (s *MAStrategy) Analyze(candles []market.Candle) Signal {
	if len(candles) < s.LongWindow+1 {
		return Hold
	}

	// Calculate MAs for the most recent candle (index len-1)
	currentShortMA := calculateMA(candles, len(candles)-1, s.ShortWindow)
	currentLongMA := calculateMA(candles, len(candles)-1, s.LongWindow)

	// Calculate MAs for the previous candle (index len-2)
	prevShortMA := calculateMA(candles, len(candles)-2, s.ShortWindow)
	prevLongMA := calculateMA(candles, len(candles)-2, s.LongWindow)

	// Cross over (short crosses above long) -> BUY
	if prevShortMA <= prevLongMA && currentShortMA > currentLongMA {
		return Buy
	}

	// Cross under (short crosses below long) -> SELL
	if prevShortMA >= prevLongMA && currentShortMA < currentLongMA {
		return Sell
	}

	return Hold
}

// calculateMA calculates the simple moving average ending at endIndex (inclusive)
func calculateMA(candles []market.Candle, endIndex int, window int) float64 {
	if endIndex-window+1 < 0 || endIndex >= len(candles) || window <= 0 {
		return 0
	}
	sum := 0.0
	for i := endIndex - window + 1; i <= endIndex; i++ {
		sum += candles[i].Close
	}
	return sum / float64(window)
}
