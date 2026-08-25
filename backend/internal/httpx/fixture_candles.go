package httpx

import (
	"math"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/market"
)

// TODO: delete once Market Data provides real historical candles.
// Synthetic sine-wave price series so real strategies (MA crossovers etc.)
// actually see price movement instead of a flat line.
func fixtureCandles(pair string, from, to int64) []market.Candle {
	const count = 120
	step := (to - from) / count
	if step < 60_000 {
		step = 60_000
	}

	candles := make([]market.Candle, count)
	for i := 0; i < count; i++ {
		price := 100 + 2*math.Sin(float64(i)/6)
		candles[i] = market.Candle{
			Symbol: pair, Timeframe: "fixture", OpenTime: from + int64(i)*step,
			Open: price, High: price + 1, Low: price - 1, Close: price + 0.5, Volume: 1,
		}
	}
	return candles
}
